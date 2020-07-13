// logx provide the standard interface of logging for what any go libraries want strip off the direct dependency to a known logging library.
package logx

type (
	Logger interface {
		Trace()
		Tracef()
		Debug()
		Debugf()
		Info()
		INfof()
		Warn()
		Warnf()
		Error()
		Errorf()
		Fatal()
		Fatalf()
		Print()
		Printf()

		SetLevel()
		GetLevel()
	}
)
