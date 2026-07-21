package mailer

import (
	"bytes"
	"fmt"
	"html/template"
)

func Render(name string, data map[string]any, siteDomain, publicAPIURL string) (subject string, html string, err error) {
	if data == nil {
		data = map[string]any{}
	}
	data["SiteDomain"] = siteDomain
	data["PublicAPIURL"] = publicAPIURL

	var t *template.Template
	switch name {
	case "verify":
		subject, t = "Verify your Frontdock email", tmplVerifyAccount
	case "reset":
		subject, t = "Reset your frontdock password", tmplResetPassword
	case "deploy_success":
		subject, t = "Deployment Succeeded", tmplDeploySuccess
	case "deploy_failure":
		subject, t = "Deployment Failed", tmplDeployFailed
	default:
		return "", "", fmt.Errorf("unknown email template %q", name)
	}
	var Buf bytes.Buffer
	if err := t.Execute(&Buf, data); err != nil {
		return "", "", err
	}
	return subject, Buf.String(), nil
}

var tmplVerifyAccount = template.Must(template.New("v").Parse(`
<div style="background:#f5f7fb;padding:40px;font-family:Arial,Helvetica,sans-serif;">
<div style="max-width:600px;margin:auto;background:white;border-radius:10px;
padding:32px;border:1px solid #e5e7eb;">

<h2 style="margin-top:0;">
Welcome 👋
</h2>

<p>
Thank you for Registering in Frontdock
Please verify your email address before continuing.
</p>

<div style="text-align:center;margin:35px 0;">

<a href="{{.PublicAPIURL}}/auth/verify?token={{.Token}}"
style="background:#2563eb;
color:white;
padding:14px 28px;
text-decoration:none;
border-radius:6px;
font-weight:bold;">
Verify Email
</a>

</div>

<p>
If the button doesn't work, copy this link into your browser:
</p>
</div>
</div>
`))

var tmplResetPassword = template.Must(template.New("r").Parse(`
<div style="background:#f5f7fb;padding:40px;font-family:Arial,Helvetica,sans-serif;">
<div style="max-width:600px;margin:auto;background:white;border-radius:10px;
padding:32px;border:1px solid #e5e7eb;">

<h2>
Reset your password
</h2>

<p>
We received a request to reset the password for your account.
</p>

<div style="text-align:center;margin:35px 0;">

<a href="{{.ResetURL}}"
style="background:#2563eb;
color:white;
padding:14px 28px;
text-decoration:none;
border-radius:6px;
font-weight:bold;">
Reset Password
</a>

</div>

<p>
If you didn't request this, you can safely ignore this email.
Your password will remain unchanged.
</p>

<p style="word-break:break-all;">
{{.ResetURL}}
</p>
</div>
</div>
`))

var tmplDeploySuccess = template.Must(template.New("deploy_success").Parse(`
<div style="background:#f5f7fb;padding:40px;font-family:Arial,Helvetica,sans-serif;">
  <div style="max-width:600px;margin:auto;background:#fff;border:1px solid #e5e7eb;border-radius:10px;padding:32px;">

    <div style="display:inline-block;background:#dcfce7;color:#166534;
        padding:6px 12px;border-radius:20px;font-size:13px;font-weight:bold;">
        ✓ Deployment Successful
    </div>

    <h2 style="margin-top:24px;color:#111827;">
        Your deployment is live 🚀
    </h2>

    <p style="color:#4b5563;">
        Hi,
    </p>

    <p style="color:#4b5563;">
        Your project
        <strong>{{.ProjectName}}</strong>
        has been successfully deployed.
    </p>

    <p style="margin:30px 0;">
        <a href="https://{{.Subdomain}}.{{.SiteDomain}}"
           style="background:#2563eb;color:white;
                  text-decoration:none;padding:12px 24px;
                  border-radius:6px;font-weight:bold;">
            Open Website
        </a>
    </p>

    <table style="width:100%;border-collapse:collapse;">
        <tr>
            <td style="padding:10px;border-bottom:1px solid #eee;">
                Version
            </td>
            <td style="padding:10px;border-bottom:1px solid #eee;">
                {{.Version}}
            </td>
        </tr>

        <tr>
            <td style="padding:10px;">
                Files Uploaded
            </td>
            <td style="padding:10px;">
                {{.FileCount}}
            </td>
        </tr>
    </table>

  </div>
</div>
`))

var tmplDeployFailed = template.Must(template.New("deploy_failed").Parse(`
<div style="background:#f5f7fb;padding:40px;font-family:Arial,Helvetica,sans-serif;">
  <div style="max-width:600px;margin:auto;background:#fff;border:1px solid #e5e7eb;border-radius:10px;padding:32px;">

    <div style="display:inline-block;background:#fee2e2;color:#991b1b;
        padding:6px 12px;border-radius:20px;font-size:13px;font-weight:bold;">
        ✕ Deployment Failed
    </div>

    <h2 style="margin-top:24px;color:#111827;">
        Deployment could not be completed
    </h2>

    <p style="color:#4b5563;">
        We couldn't deploy
        <strong>{{.ProjectName}}</strong>.
    </p>

    <table style="width:100%;margin-top:20px;border-collapse:collapse;">
        <tr>
            <td style="padding:10px;border-bottom:1px solid #eee;">
                Version
            </td>
            <td style="padding:10px;border-bottom:1px solid #eee;">
                {{.Version}}
            </td>
        </tr>

        <tr>
            <td style="padding:10px;">
                Reason
            </td>
            <td style="padding:10px;color:#dc2626;">
                {{.Reason}}
            </td>
        </tr>
    </table>

    <p style="margin-top:28px;color:#6b7280;">
        Your currently deployed version is still online and serving visitors normally.
    </p>

  </div>
</div>
`))
