package auth

var loginPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Login - BandCash</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <main id="app" class="single-col">
        <h1>Login to BandCash</h1>
        <nav class="row pb"><a href="/">Home</a><span>&gt;</span><span>Login</span></nav>
        <form method="POST" action="/auth/login">
            <div class="field">
                <label>Email</label>
                <input type="email" name="email" required placeholder="your@email.com">
            </div>
            <button type="submit" class="btn btn-primary">Send Login Link</button>
        </form>
        <p>Don't have an account? <a href="/auth/signup">Sign up</a></p>
        <p><a href="/">← Back to home</a></p>
    </main>
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
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <main id="app" class="single-col">
        <h1>Create Account</h1>
        <nav class="row pb"><a href="/">Home</a><span>&gt;</span><span>Sign Up</span></nav>
        <form method="POST" action="/auth/signup">
            <div class="field">
                <label>Email</label>
                <input type="email" name="email" required placeholder="your@email.com" ` + emailValue + `>
            </div>
            <button type="submit" class="btn btn-primary">Create Account</button>
        </form>
        <p>Already have an account? <a href="/auth/login">Login</a></p>
        <p><a href="/">← Back to home</a></p>
    </main>
</body>
</html>`
}

var loginSentPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Check Your Email - BandCash</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <main id="app" class="single-col">
        <h1>Check Your Email</h1>
        <nav class="row pb"><a href="/">Home</a><span>&gt;</span><span>Check Email</span></nav>
        <p>We've sent you a login link. Click the link in your email to access your account.</p>
        <p>The link will expire in 1 hour.</p>
        <p><a href="/">← Back to home</a></p>
    </main>
</body>
</html>`
