package model

type Thumbnail struct {
	width   int
	height  int
	preview []byte
}

func NewThumbnail(w int, h int) *Thumbnail {
	return &Thumbnail{
		width:   w,
		height:  h,
		preview: nil,
	}
}
