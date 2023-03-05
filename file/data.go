package file

import (
	"os"

	logger "github.com/wawakakakyakya/GolangLogger"
	"github.com/wawakakakyakya/check_logs_by_mail/config"
)

type File struct {
	PosFile *PosFile
	FileReadWriter
	Config *config.FileConfig
	logger *logger.Logger
}

func (f *File) IsRotated(path string) (bool, error) {
	var inode uint64
	var err error
	if inode, err = f.Inode(path); err != nil {
		f.logger.ErrorF("get inode failed: %s", path)
		return false, err
	}
	f.logger.DebugF("posInode: %d, orgInode(%s): %d", f.PosFile.Inode, path, inode)
	return f.PosFile.Inode != inode, nil
}

func (f *File) Parse(path string, logparser *LogParser) (int64, bool, error) {
	var readSize int
	var rotated bool
	var err error

	if f.PosFile.Inode != 0 {
		rotated, err = f.IsRotated(path)
		if err != nil {
			return 0, false, err
		}
	}

	// 初回はローテーションファイルは処理しない
	if rotated {
		rotateFile, err := f.GetOneGenerationBeforeFile(path, f.PosFile.Inode)
		if err != nil {
			return 0, false, err
		}
		//ローテーションされたファイルなので、ポジションには含めない
		f.logger.DebugF("%s was rorated, load old file: %s", path, rotateFile)
		_, _, err = f.Parse(rotateFile, logparser)
		if err != nil {
			return 0, false, err
		}
		f.logger.InfoF("rotated file(%s) readed", rotateFile)
	}

	f.logger.InfoF("read file(%s) started", path)
	fp, err := os.Open(path)
	if err != nil {
		return 0, false, err
	}
	defer fp.Close()

	fp.Seek(f.PosFile.LastLine, 0)
	f.logger.DebugF("move to lastline(%d) in %s", f.PosFile.LastLine, path)
	readSize, err = logparser.Parse(fp, path, f.Config.Words)

	return int64(readSize), rotated, err
}

func (f *File) UpdatePosition(lastLine int64) error {
	inode, err := f.Inode(f.Config.FileName)
	if err != nil {
		return err
	}
	if err := f.PosFile.Write(inode, lastLine); err != nil {
		f.logger.ErrorF("update position failed: %s", f.PosFile.FileName)
		return err
	}
	return nil
}

func NewFile(config *config.FileConfig, logger *logger.Logger) (*File, error) {
	fileLogger := logger.Child("file")
	posFile, err := NewPosFile(config.PosFile, fileLogger)
	if err != nil {
		logger.ErrorF("new posfile failed")
		return nil, err
	}

	return &File{PosFile: posFile, Config: config, logger: fileLogger}, nil
}
