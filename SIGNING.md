# Code Signing Setup for RambleAI

## Local Signing (Recommended)

This approach signs your app locally without dealing with CI/CD secrets or expired passwords.

### Prerequisites

1. **Apple Developer Account** with "Developer ID Application" certificate
2. **Certificate installed** in your macOS Keychain

### Setup

1. **Check your certificates** (the Makefile will auto-detect):
   ```bash
   make check-signing
   ```

2. **Optional**: Set specific certificate (if you have multiple):
   ```bash
   export APPLE_DEVELOPER_ID="Developer ID Application: Your Name (TEAM123456)"
   ```

3. **Optional**: For notarization, set Apple ID credentials:
   ```bash
   export APPLE_ID="your-apple-id@email.com"
   export APPLE_ID_PASSWORD="your-app-specific-password"  
   export APPLE_TEAM_ID="TEAM123456"
   ```

### Usage

```bash
# Check your signing setup
make check-signing

# Build and sign in one command (auto-detects certificate)
make build-and-sign

# Or build first, then sign
make build
make sign

# Create signed zip for distribution (with notarization)
make sign-zip
```

### Environment Variables

Add these to your `~/.zshrc` or `~/.bash_profile`:

```bash
# Required for signing
export APPLE_DEVELOPER_ID="Developer ID Application: Your Name (TEAM123456)"

# Optional - for notarization
export APPLE_ID="your-apple-id@email.com" 
export APPLE_ID_PASSWORD="xxxx-xxxx-xxxx-xxxx"  # App-specific password
export APPLE_TEAM_ID="TEAM123456"
```

### Manual Signing

You can also sign manually using the script:

```bash
# Sign the app bundle
./script/sign build/bin/RambleAI.app

# Sign a zip for notarization  
./script/sign build/RambleAI-macos.zip
```

## Troubleshooting

### "No signing certificate found"
- Check: `security find-identity -v -p codesigning`
- Make sure you have a "Developer ID Application" certificate (not "Mac App Distribution")

### "App can't be opened because Apple cannot check it"
- The app wasn't signed, or 
- You're running a build from someone else's machine
- Re-sign with: `make sign`

### App-specific password issues
- Generate new one at: https://appleid.apple.com → App-Specific Passwords
- They expire after 1 year (yes, it's annoying)

## Benefits of Local Signing

✅ **No CI secrets to manage**  
✅ **No expired passwords breaking builds**  
✅ **Faster feedback loop**  
✅ **Works offline**  
✅ **More reliable than GitHub Actions**