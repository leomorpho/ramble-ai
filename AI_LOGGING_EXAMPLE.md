# AI Endpoint Logging Example

With the enhanced logging in place, every AI request to PocketBase will generate comprehensive logs. Here's what you'll see:

## Text Processing Request (AI Reordering, Suggestions, etc.)

```bash
# Request starts
🤖 [AI TEXT REQUEST] IP: 192.168.1.100 | User-Agent: Mozilla/5.0... | Method: POST

# API key validation
🔐 [AI TEXT REQUEST] API Key: ra-abc123... | IP: 192.168.1.100

# User identification
👤 [AI TEXT REQUEST] User: user@example.com (rec_xyz789) | API Key: ra-abc123... | IP: 192.168.1.100

# Request processing details
📝 [AI TEXT REQUEST] Processing | User: user@example.com | Task: reorder | Model: anthropic/claude-3.5-sonnet | Prompt Length: 1234 chars | System Prompt Length: 567 chars | IP: 192.168.1.100

# Usage analytics
📊 [AI USAGE] User: user@example.com (rec_xyz789) | Task: reorder | Model: anthropic/claude-3.5-sonnet | Input: 1234 | Output: 2345 | Duration: 3.2s | IP: 192.168.1.100

# Success confirmation
✅ [AI TEXT REQUEST] SUCCESS | User: user@example.com | Task: reorder | Model: anthropic/claude-3.5-sonnet | Response Length: 2345 chars | Duration: 3.2s | IP: 192.168.1.100
```

## Audio Transcription Request

```bash
# Request starts
🎵 [AI AUDIO REQUEST] IP: 192.168.1.100 | User-Agent: RambleAI/1.0... | Method: POST

# API key validation
🔐 [AI AUDIO REQUEST] API Key: ra-abc123... | IP: 192.168.1.100

# User identification  
👤 [AI AUDIO REQUEST] User: user@example.com (rec_xyz789) | API Key: ra-abc123... | IP: 192.168.1.100

# Audio processing details
🎵 [AI AUDIO REQUEST] Processing | User: user@example.com | Filename: recording.wav | Audio Size: 1024 KB | IP: 192.168.1.100

# Usage analytics
📊 [AI USAGE] User: user@example.com (rec_xyz789) | Task: transcription | Model: whisper-1 | Input: 1024 | Output: 3456 | Duration: 15.7s | IP: 192.168.1.100

# Success confirmation
✅ [AI AUDIO REQUEST] SUCCESS | User: user@example.com | Filename: recording.wav | Audio: 1024 KB | Transcript: 3456 chars | Words: 432 | Duration: 15.7s | IP: 192.168.1.100
```

## API Key Generation

```bash
# Request starts
🔑 [API KEY REQUEST] IP: 192.168.1.100 | User-Agent: Mozilla/5.0...

# User identification
👤 [API KEY REQUEST] User: user@example.com (rec_xyz789) | IP: 192.168.1.100

# Success confirmation
✅ [API KEY REQUEST] SUCCESS: Generated API key ra-def456... | User: user@example.com | IP: 192.168.1.100
```

## Error Cases

```bash
# Missing API key
❌ [AI TEXT REQUEST] FAILED: Missing API key | IP: 192.168.1.100

# Invalid API key
❌ [AI TEXT REQUEST] FAILED: Invalid API key ra-invalid... | IP: 192.168.1.100 | Error: API key not found or inactive

# No subscription
❌ [AI TEXT REQUEST] FAILED: No active subscription | User: user@example.com | IP: 192.168.1.100

# OpenRouter error
❌ [AI TEXT REQUEST] FAILED: OpenRouter error | User: user@example.com | Task: reorder | Model: anthropic/claude-3.5-sonnet | Duration: 1.2s | IP: 192.168.1.100 | Error: rate limit exceeded
```

## Key Features

✅ **Request Tracking**: Every request is logged with timestamp, IP, and user agent
✅ **User Attribution**: All requests are tied to specific users and email addresses  
✅ **API Key Security**: Keys are masked in logs (only first 8 characters shown)
✅ **Performance Monitoring**: Response times and processing duration tracked
✅ **Usage Analytics**: Input/output sizes, model usage, task types tracked
✅ **Error Logging**: Detailed error information for debugging
✅ **Security Insights**: IP addresses and failed authentication attempts logged

## Analytics Potential

The logs provide rich data for:
- **Usage Billing**: Track API calls per user, model usage, processing time
- **Performance Monitoring**: Identify slow requests, API issues
- **Security Monitoring**: Detect suspicious activity, failed auth attempts
- **User Behavior**: Understand which AI features are most popular
- **Cost Analysis**: Track OpenRouter/OpenAI API costs per user

## Database Storage (Optional)

The code includes commented-out database storage functionality. To enable persistent logging, you could create an `ai_usage_logs` collection in PocketBase with fields:
- user_id
- task_type  
- model
- tokens_used
- input_size
- output_size
- duration_ms
- client_ip
- timestamp

This would enable building dashboards, billing systems, and detailed analytics.