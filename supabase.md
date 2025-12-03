1. Create Account

Go to supabase.com
Sign up with GitHub (easiest)

2. Create Project

Click New Project
Select your organization
Fill in:

Name: 247techspyware-blog (or whatever)
Database Password: Generate a strong one and copy it somewhere safe - you won't see it again
Region: Choose closest to your users (e.g., London, Frankfurt for Nigeria)


Click Create new project
Wait 2-3 minutes for setup

3. Get Your Credentials

Once ready, go to Settings (gear icon left sidebar)
Click Database
Scroll down to Connection parameters section

You'll see:
Host:     db.xxxxxxxxxxxx.supabase.co
Database: postgres
Port:     5432
User:     postgres
Password: [the one you set]
4. Your .env file
envDB_TYPE=postgres
DB_HOST=db.xxxxxxxxxxxx.supabase.co
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=the-password-you-saved
DB_NAME=postgres
DB_SSLMODE=require  




git remote add origin https://github.com/IMAdegboyega/247spywareAPI.git
git branch -M main
git push -u origin main