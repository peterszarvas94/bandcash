package auth

var loginPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Login - BandCash</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <nav class="row pb"><a href="/">Home</a><span>&gt;</span><span>Login</span></nav>
        <h1>Login to BandCash</h1>
        <form method="POST" action="/auth/login">
            <div class="field">
                <label>Email</label>
                <input type="email" name="email" required placeholder="your@email.com">
            </div>
            <button type="submit" class="btn btn-primary">Send Login Link</button>
        </form>
        <p>Don't have an account? <a href="/auth/signup">Sign up</a></p>
        <p><a href="/">← Back to home</a></p>
    </div>
</body>
</html>`

func signupPageHTML(email string) string {
	emailValue := ""
	if email != "" {
		emailValue = `value="` + email + `"`
	}
	return `<!DOCTYPE html>
<html>
<head>
    <title>Sign Up - BandCash</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <nav class="row pb"><a href="/">Home</a><span>&gt;</span><span>Sign Up</span></nav>
        <h1>Create Account</h1>
        <form method="POST" action="/auth/signup">
            <div class="field">
                <label>Email</label>
                <input type="email" name="email" required placeholder="your@email.com" ` + emailValue + `>
            </div>
            <button type="submit" class="btn btn-primary">Create Account</button>
        </form>
        <p>Already have an account? <a href="/auth/login">Login</a></p>
        <p><a href="/">← Back to home</a></p>
    </div>
</body>
</html>`
}

var loginSentPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Check Your Email - BandCash</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <nav class="row pb"><a href="/">Home</a><span>&gt;</span><span>Check Email</span></nav>
        <h1>Check Your Email</h1>
        <p>We've sent you a login link. Click the link in your email to access your account.</p>
        <p>The link will expire in 1 hour.</p>
        <p><a href="/">← Back to home</a></p>
    </div>
</body>
</html>`
