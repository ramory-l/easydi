package widget

// Widget is produced by NewWidget.
type Widget struct{}

// di:provide name=Widget
// di:expose
func NewWidget(
	// di:use Svc
	l interface{ Get() string },
) *Widget {
	return &Widget{}
}
