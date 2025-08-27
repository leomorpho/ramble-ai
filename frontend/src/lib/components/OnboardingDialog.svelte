<script>
  import { 
    Dialog, 
    DialogContent, 
    DialogDescription, 
    DialogHeader, 
    DialogTitle, 
  } from "$lib/components/ui/dialog";
  import { Button } from "$lib/components/ui/button";
  import { ExternalLink, Rocket } from "@lucide/svelte";

  let { open = $bindable(false) } = $props();
</script>

<Dialog bind:open>
  <DialogContent class="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
    <DialogHeader>
      <DialogTitle class="text-xl flex items-center gap-2">
        <Rocket class="w-5 h-5 text-primary" />
        Welcome to RambleAI!
      </DialogTitle>
      <DialogDescription>
        Get started in just a few minutes. Set up your API keys to unlock powerful AI features.
      </DialogDescription>
    </DialogHeader>
    
    <div class="space-y-6 py-4">
      <!-- OpenAI Section -->
      <div class="bg-secondary/30 border border-border rounded-lg p-4 space-y-4">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-green-500/10 rounded-full flex items-center justify-center">
            <span class="text-green-600 font-bold text-sm">1</span>
          </div>
          <h3 class="text-lg font-semibold">Set up OpenAI (Required)</h3>
        </div>
        
        <p class="text-sm text-muted-foreground pl-10">
          <strong>What you'll get:</strong> Automatic audio transcription and speech-to-text conversion for your videos. <strong>Required for the app to function.</strong>
        </p>
        
        <div class="pl-10 space-y-2">
          <div class="text-sm space-y-1">
            <p class="font-medium">Steps:</p>
            <ol class="text-muted-foreground space-y-1 ml-4 list-decimal">
              <li>
                Go to <a href="https://platform.openai.com/api-keys" target="_blank" class="text-primary hover:underline font-medium inline-flex items-center gap-1">
                  platform.openai.com/api-keys <ExternalLink class="w-3 h-3" />
                </a>
              </li>
              <li>Sign in or create an OpenAI account</li>
              <li>Click "Create new secret key" and give it a name</li>
              <li>Copy the key (starts with "sk-") and add it in Settings</li>
            </ol>
          </div>
          <div class="text-xs text-muted-foreground bg-secondary/50 p-2 rounded">
            <strong>💰 Cost:</strong> ~$0.006 per minute of audio transcription.
          </div>
        </div>
      </div>

      <!-- OpenRouter Section -->
      <div class="bg-secondary/30 border border-border rounded-lg p-4 space-y-4">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-blue-500/10 rounded-full flex items-center justify-center">
            <span class="text-blue-600 font-bold text-sm">2</span>
          </div>
          <h3 class="text-lg font-semibold">Set up OpenRouter (Optional)</h3>
        </div>
        
        <p class="text-sm text-muted-foreground pl-10">
          <strong>What you'll get:</strong> Intelligent reordering of video segments, content-aware highlight organization, and AI-powered editing suggestions.
        </p>
        
        <div class="pl-10 space-y-2">
          <div class="text-sm space-y-1">
            <p class="font-medium">Steps:</p>
            <ol class="text-muted-foreground space-y-1 ml-4 list-decimal">
              <li>
                Go to <a href="https://openrouter.ai/keys" target="_blank" class="text-primary hover:underline font-medium inline-flex items-center gap-1">
                  openrouter.ai/keys <ExternalLink class="w-3 h-3" />
                </a>
              </li>
              <li>Sign in with Google, GitHub, or create an account</li>
              <li>Click "Create Key" and give it a name</li>
              <li>Copy the key (starts with "sk-or-") and add it in Settings</li>
            </ol>
          </div>
          <div class="text-xs text-muted-foreground bg-secondary/50 p-2 rounded">
            <strong>ℹ️ Note:</strong> Check OpenRouter pricing for current rates.
          </div>
        </div>
      </div>

      <!-- Quick Start -->
      <div class="bg-primary/5 border border-primary/20 rounded-lg p-4 space-y-3">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
            <span class="text-primary font-bold text-sm">3</span>
          </div>
          <h3 class="text-lg font-semibold">Start Creating!</h3>
        </div>
        
        <div class="pl-10 text-sm text-muted-foreground space-y-1">
          <p>Once you've set up your API keys:</p>
          <ul class="ml-4 space-y-1 list-disc">
            <li>Create your first project</li>
            <li>Upload or drop video files</li>
            <li>Let AI transcribe and organize your content</li>
            <li>Export polished videos and transcripts</li>
          </ul>
        </div>
      </div>
    </div>

    <div class="flex gap-3 pt-4">
      <Button asChild class="flex-1">
        <a href="/settings">Go to Settings</a>
      </Button>
      <Button variant="outline" onclick={() => open = false} disabled={false} class="flex-1">
        I'll Set This Up Later
      </Button>
    </div>
  </DialogContent>
</Dialog>