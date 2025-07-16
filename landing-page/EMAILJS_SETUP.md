# EmailJS Setup Instructions

To enable email sending from your static frontend, follow these steps to set up EmailJS:

## 1. Create EmailJS Account
1. Go to [EmailJS.com](https://www.emailjs.com/)
2. Sign up for a free account
3. Verify your email address

## 2. Set Up Email Service
1. In your EmailJS dashboard, go to "Email Services"
2. Click "Add New Service"
3. Choose your email provider (Gmail, Outlook, Yahoo, etc.)
4. Follow the setup instructions for your provider
5. Note your **Service ID**

## 3. Create Email Template
1. Go to "Email Templates" in your dashboard
2. Click "Create New Template"
3. Use this template structure:

```
Subject: New Project Inquiry from {{from_name}}

Contact Information:
Name: {{from_name}}
Email: {{from_email}}
Company: {{company}}
Phone: {{phone}}

Project Details:
Type: {{project_type}}
Size: {{project_size}}
Budget: {{budget}}
Timeline: {{timeline}}

Description:
{{description}}

Technical Requirements:
Platforms: {{platforms}}
Existing Technology: {{existing_tech}}
Integrations: {{integrations}}

Additional Information:
Inspiration: {{inspiration}}
Notes: {{additional_notes}}

Submitted: {{submission_date}}
```

4. Note your **Template ID**

## 4. Get Your User ID
1. Go to "Account" in your dashboard
2. Find your **User ID** (also called Public Key)

## 5. Install EmailJS
```bash
npm install @emailjs/browser
```

## 6. Update the Component

Replace the placeholder in `ProjectIntakeForm.svelte`:

```typescript
// Add this import at the top
import emailjs from '@emailjs/browser';

// Replace the TODO section in handleSubmit with:
await emailjs.send(
  'YOUR_SERVICE_ID',    // Replace with your Service ID
  'YOUR_TEMPLATE_ID',   // Replace with your Template ID
  emailData,
  'YOUR_USER_ID'        // Replace with your User ID
);
```

## 7. Environment Variables (Optional)
For security, you can store these in environment variables:

```typescript
// In your .env file
VITE_EMAILJS_SERVICE_ID=your_service_id
VITE_EMAILJS_TEMPLATE_ID=your_template_id
VITE_EMAILJS_USER_ID=your_user_id

// In your component
await emailjs.send(
  import.meta.env.VITE_EMAILJS_SERVICE_ID,
  import.meta.env.VITE_EMAILJS_TEMPLATE_ID,
  emailData,
  import.meta.env.VITE_EMAILJS_USER_ID
);
```

## 8. Test Your Setup
1. Fill out the form and submit
2. Check your email inbox
3. Verify all data is being sent correctly

## Alternative Services
If you prefer other services:
- **Formspree**: Simple form backend
- **Netlify Forms**: If hosting on Netlify
- **Getform**: Form backend service
- **Formspark**: Lightweight form backend

## Benefits of This Setup?
✅ **Fully static** - No backend required
✅ **Free tier available** - 200 emails/month on EmailJS free plan
✅ **Client-side persistence** - Form data saved in localStorage
✅ **Professional emails** - Formatted and comprehensive
✅ **Easy setup** - No server configuration needed