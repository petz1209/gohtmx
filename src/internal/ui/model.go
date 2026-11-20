package ui

type PageData struct {
	Title  string
	Header Header
	Footer Footer
	Data   any
}

type PageError struct {
	Title      string
	Header     Header
	Footer     Footer
	StatusCode int
	Error      string
}

type Header struct {
}

type Footer struct {
}
