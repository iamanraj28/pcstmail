package pcstmail_test

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"time"

	"github.com/iamanraj28/pcstmail"
)

func Example() {
	m := pcstmail.NewMessage()
	m.SetHeader("From", "aman@example.com")
	m.SetHeader("To", "ami@example.com", "ame@example.com")
	m.SetAddressHeader("Cc", "ama@example.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	m.Attach("/home/Alex/lolcat.jpg")

	d := pcstmail.NewDialer("smtp.example.com", 587, "user", "123456")

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

// A daemon that listens to a channel and sends all incoming messages.
func Example_daemon() {
	ch := make(chan *pcstmail.Message)

	go func() {
		d := pcstmail.NewDialer("smtp.example.com", 587, "user", "123456")

		var s pcstmail.SendCloser
		var err error
		open := false
		for {
			select {
			case m, ok := <-ch:
				if !ok {
					return
				}
				if !open {
					if s, err = d.Dial(); err != nil {
						panic(err)
					}
					open = true
				}
				if err := pcstmail.Send(s, m); err != nil {
					log.Print(err)
				}
			// Close the connection to the SMTP server if no email was sent in
			// the last 30 seconds.
			case <-time.After(30 * time.Second):
				if open {
					if err := s.Close(); err != nil {
						panic(err)
					}
					open = false
				}
			}
		}
	}()

	// Use the channel in your program to send emails.

	// Close the channel to stop the mail daemon.
	close(ch)
}

// Efficiently send a customized newsletter to a list of recipients.
func Example_newsletter() {
	// The list of recipients.
	var list []struct {
		Name    string
		Address string
	}

	d := pcstmail.NewDialer("smtp.example.com", 587, "user", "123456")
	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	m := pcstmail.NewMessage()
	for _, r := range list {
		m.SetHeader("From", "no-reply@example.com")
		m.SetAddressHeader("To", r.Address, r.Name)
		m.SetHeader("Subject", "Newsletter #1")
		m.SetBody("text/html", fmt.Sprintf("Hello %s!", r.Name))

		if err := pcstmail.Send(s, m); err != nil {
			log.Printf("Could not send email to %q: %v", r.Address, err)
		}
		m.Reset()
	}
}

// Send an email using a local SMTP server.
func Example_noAuth() {
	m := pcstmail.NewMessage()
	m.SetHeader("From", "from@example.com")
	m.SetHeader("To", "to@example.com")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/plain", "Hello!")

	d := pcstmail.Dialer{Host: "localhost", Port: 587}
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

// Send an email using an API or postfix.
func Example_noSMTP() {
	m := pcstmail.NewMessage()
	m.SetHeader("From", "from@example.com")
	m.SetHeader("To", "to@example.com")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/plain", "Hello!")

	s := pcstmail.SendFunc(func(from string, to []string, msg io.WriterTo) error {
		// Implements you email-sending function, for example by calling
		// an API, or running postfix, etc.
		fmt.Println("From:", from)
		fmt.Println("To:", to)
		return nil
	})

	if err := pcstmail.Send(s, m); err != nil {
		panic(err)
	}
	// Output:
	// From: from@example.com
	// To: [to@example.com]
}

var m *pcstmail.Message

func ExampleSetCopyFunc() {
	m.Attach("foo.txt", pcstmail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write([]byte("Content of foo.txt"))
		return err
	}))
}

func ExampleSetHeader() {
	h := map[string][]string{"Content-ID": {"<foo@bar.mail>"}}
	m.Attach("foo.jpg", pcstmail.SetHeader(h))
}

func ExampleRename() {
	m.Attach("/tmp/0000146.jpg", pcstmail.Rename("picture.jpg"))
}

func ExampleMessage_AddAlternative() {
	m.SetBody("text/plain", "Hello!")
	m.AddAlternative("text/html", "<p>Hello!</p>")
}

func ExampleMessage_AddAlternativeWriter() {
	t := template.Must(template.New("example").Parse("Hello {{.}}!"))
	m.AddAlternativeWriter("text/plain", func(w io.Writer) error {
		return t.Execute(w, "Bob")
	})
}

func ExampleMessage_Attach() {
	m.Attach("/tmp/image.jpg")
}

func ExampleMessage_Embed() {
	m.Embed("/tmp/image.jpg")
	m.SetBody("text/html", `<img src="cid:image.jpg" alt="My image" />`)
}

func ExampleMessage_FormatAddress() {
	m.SetHeader("To", m.FormatAddress("ami@example.com", "Bob"), m.FormatAddress("ame@example.com", "Cora"))
}

func ExampleMessage_FormatDate() {
	m.SetHeaders(map[string][]string{
		"X-Date": {m.FormatDate(time.Now())},
	})
}

func ExampleMessage_SetAddressHeader() {
	m.SetAddressHeader("To", "ami@example.com", "Bob")
}

func ExampleMessage_SetBody() {
	m.SetBody("text/plain", "Hello!")
}

func ExampleMessage_SetDateHeader() {
	m.SetDateHeader("X-Date", time.Now())
}

func ExampleMessage_SetHeader() {
	m.SetHeader("Subject", "Hello!")
}

func ExampleMessage_SetHeaders() {
	m.SetHeaders(map[string][]string{
		"From":    {m.FormatAddress("aman@example.com", "Alex")},
		"To":      {"ami@example.com", "ame@example.com"},
		"Subject": {"Hello"},
	})
}

func ExampleSetCharset() {
	m = pcstmail.NewMessage(pcstmail.SetCharset("ISO-8859-1"))
}

func ExampleSetEncoding() {
	m = pcstmail.NewMessage(pcstmail.SetEncoding(pcstmail.Base64))
}

func ExampleSetPartEncoding() {
	m.SetBody("text/plain", "Hello!", pcstmail.SetPartEncoding(pcstmail.Unencoded))
}
