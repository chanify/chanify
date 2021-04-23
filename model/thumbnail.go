package model

// Thumbnail for image
type Thumbnail struct {
	width   int
	height  int
	preview []byte
}

// NewThumbnail from image width & height
func NewThumbnail(w int, h int) *Thumbnail {
	return &Thumbnail{
		width:   w,
		height:  h,
		preview: nil,
	}
}
