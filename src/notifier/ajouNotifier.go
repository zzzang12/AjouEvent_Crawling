package notifier

type Notifier interface {
	Notify()
}

type NotifierStruct struct {
	notifier Notifier
}

type Notice struct {
	ID         string
	Category   string
	Title      string
	Department string
	Date       string
	Link       string
}

type BaseNotifier struct {
	BoxCount  int
	MaxNum    int
	URL       string
	ChannelID string
	FsDocID   string
}

func (NotifierStruct) New(source Notifier) *NotifierStruct {
	return &NotifierStruct{source}
}

func (notifier *NotifierStruct) Notify() {
	notifier.notifier.Notify()
}

func CreateNotifier(notifierName string) *NotifierStruct {
	switch notifierName {
	case "AjouNormal":
		return NotifierStruct{}.New(AjouNormalNotifier{}.New())
	case "AjouScholarship":
		return NotifierStruct{}.New(AjouScholarshipNotifier{}.New())
	case "AjouSw":
		return NotifierStruct{}.New(AjouSwNotifier{}.New())
	case "AjouSoftware":
		return NotifierStruct{}.New(AjouSoftwareNotifier{}.New())
	default:
		return nil
	}
}
