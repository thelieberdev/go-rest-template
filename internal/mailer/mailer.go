package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"os"

	"github.com/wneessen/go-mail"
)

//go:embed "templates"
var templateFS embed.FS

type Config struct {
	Host     	string
	Port     	int
	Username 	string
	Password 	string
	Sender   	string
}

type Mailer struct {
	client  	 *mail.Client
	sender  	 string
	errLogger  *slog.Logger
}

func Init(cfg Config, errLogger *slog.Logger) (*Mailer, error) {
	client, err := mail.NewClient(
		cfg.Host,
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover), 
		mail.WithPort(cfg.Port),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
	)
	if err != nil { 
		return nil, err 
	}

	return &Mailer{ client: client, sender: cfg.Sender, }, nil
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From(m.sender); err != nil {
    m.errLogger.Error("failed to set FROM address: " + err.Error())
		os.Exit(1)
	}
	if err := msg.To(recipient); err != nil {
		m.errLogger.Error("failed to set TO address: " + err.Error())
		os.Exit(1)
	}
	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	if err := m.client.DialAndSend(msg); err != nil {
		m.errLogger.Error("failed to deliver mail: " + err.Error())
		os.Exit(1)
	}

	return err
}
