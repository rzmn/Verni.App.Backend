package pathProvider

import "path/filepath"

type defaultService struct {
	root string
}

func (c *defaultService) AbsolutePath(relative string) string {
	return filepath.Join(c.root, relative)
}
