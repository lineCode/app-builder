package fs

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/apex/log"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
	"github.com/develar/go-fs-util"
)

type FileCopier struct {
	IsUseHardLinks bool

	links []LinkInfo
}

type LinkInfo struct {
	File string
	Link string
}

// go doesn't provide native copy operation (CoW)
func (t *FileCopier) copyDir(from string, to string) error {
	fileNames, err := fsutil.ReadDirContent(from)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, name := range fileNames {
		if name == ".DS_Store" {
			continue
		}

		err = t.copyDirOrFile(filepath.Join(from, name), filepath.Join(to, name), false)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func CopyUsingHardlink(from string, to string) error {
	var fileCopier FileCopier
	fileCopier.IsUseHardLinks = true
	return fileCopier.CopyDirOrFile(from, to)
}

func CopyDirOrFile(from string, to string) error {
	var fileCopier FileCopier
	return fileCopier.CopyDirOrFile(from, to)
}

func (t *FileCopier) CopyDirOrFile(from string, to string) error {
	if runtime.GOOS == "windows" {
		t.IsUseHardLinks = false
	}

	log.WithFields(log.Fields{
		"from":           from,
		"to":             to,
		"isUseHardLinks": t.IsUseHardLinks,
	}).Debug("copy files")

	err := t.copyDirOrFile(from, to, true)
	if err != nil {
		return errors.WithStack(err)
	}

	if t.links != nil {
		err = CreateLinks(t.links)
		if err != nil {
			return errors.WithStack(err)
		}
		t.links = nil
	}

	return nil
}

func CreateLinks(links []LinkInfo) error {
	return util.MapAsync(len(links), func(taskIndex int) (func() error, error) {
		linkInfo := links[taskIndex]
		return func() error {
			err := os.Symlink(linkInfo.Link, linkInfo.File)
			if err != nil {
				return err
			}
			return nil
		}, nil
	})
}

func (t *FileCopier) copyDirOrFile(from string, to string, isCreateParentDirs bool) error {
	fromInfo, err := os.Lstat(from)
	if err != nil {
		return errors.WithStack(err)
	}

	if fromInfo.IsDir() {
		// cannot use file mode as is because of *** *** *** umask
		if isCreateParentDirs {
			err = fsutil.EnsureDir(to)
		} else {
			err = os.Mkdir(to, 0777)
		}

		err = SetDirPermsIfNeed(to, fromInfo.Mode())
		if err != nil {
			return errors.WithStack(err)
		}

		if err != nil && !os.IsExist(err) {
			return errors.WithStack(err)
		}

		return t.copyDir(from, to)
	} else if fromInfo.Mode()&os.ModeSymlink != 0 {
		return t.copySymlink(from, to)
	}

	return t.CopyFile(from, to, isCreateParentDirs, fromInfo)
}

func SetDirPermsIfNeed(dir string, mode os.FileMode) error {
	perm := mode.Perm()
	if perm != 0755 {
		err := os.Chmod(dir, perm)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (t *FileCopier) CopyFile(from string, to string, isCreateParentDirs bool, fromInfo os.FileInfo) error {
	if t.IsUseHardLinks {
		err := os.Link(from, to)
		if err == nil {
			return nil
		}

		t.IsUseHardLinks = false
		log.WithError(err).WithField("from", from).WithField("to", to).Debug("cannot copy using hard link")
	}

	if isCreateParentDirs {
		err := fsutil.EnsureDir(filepath.Dir(to))
		if err != nil {
			return err
		}
	}
	return fsutil.CopyFile(from, to, fromInfo)
}

// symlink cannot be created during copy because symlink can point to not yet copied target file
func (t *FileCopier) copySymlink(from string, to string) error {
	link, err := os.Readlink(from)
	if err != nil {
		return errors.WithStack(err)
	}

	if filepath.IsAbs(link) {
		link, err = filepath.Rel(filepath.Dir(from), link)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	t.links = append(t.links, LinkInfo{
		File: to,
		Link: link,
	})

	return nil
}
