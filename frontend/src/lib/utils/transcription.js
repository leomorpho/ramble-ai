import { 
  GetOpenAIApiKey, 
  TranscribeVideoClip, 
  UpdateVideoClip 
} from '$lib/wailsjs/go/main/App.js';
import { toast } from 'svelte-sonner';

export const TranscriptionState = {
  IDLE: 'idle',
  CHECKING: 'checking',
  TRANSCRIBING: 'transcribing',
  COMPLETED: 'completed',
  ERROR: 'error'
};

/**
 * Start transcription for a video clip
 * @param {Object} clip - The video clip object
 * @returns {Promise<Object>} Updated video clip with transcription
 */
export async function startTranscription(clip) {
  if (!clip || !clip.id) {
    throw new Error('Invalid clip provided');
  }

  // Check if OpenAI API key is configured
  const apiKey = await GetOpenAIApiKey();
  if (!apiKey || apiKey.trim() === '') {
    toast.error('OpenAI API key not configured', {
      description: 'Please configure your OpenAI API key in settings',
      action: {
        label: 'Open Settings',
        onClick: () => window.location.href = '/settings'
      }
    });
    throw new Error('OpenAI API key not configured');
  }

  try {
    // Start transcription - no need to update state here since it's handled by the caller
    const result = await TranscribeVideoClip(clip.id);
    
    toast.success('Transcription completed', {
      description: `Detected language: ${result.language || 'Unknown'}`
    });
    
    return result;
  } catch (error) {
    console.error('Transcription error:', error);
    
    toast.error('Failed to transcribe video', {
      description: error.message || 'Unknown error occurred'
    });
    
    throw error;
  }
}

/**
 * Update transcription state for a video clip
 * @param {string} clipId - The video clip ID
 * @param {string} state - The transcription state
 * @param {string} error - Optional error message
 * @returns {Promise<void>}
 */
async function updateClipTranscriptionState(clipId, state, error = '') {
  try {
    // Since we don't have a direct API for updating transcription state,
    // we'll use the general UpdateVideoClip function
    await UpdateVideoClip(clipId, {
      transcriptionState: state,
      transcriptionError: error,
      transcriptionStartedAt: state === TranscriptionState.TRANSCRIBING ? new Date() : undefined,
      transcriptionCompletedAt: state === TranscriptionState.COMPLETED ? new Date() : undefined
    });
  } catch (err) {
    console.error('Failed to update transcription state:', err);
  }
}

/**
 * Get the current transcription state for a clip
 * @param {Object} clip - The video clip object
 * @returns {string} The transcription state
 */
export function getTranscriptionState(clip) {
  if (!clip) return TranscriptionState.IDLE;
  
  // Use the new transcription_state field if available
  if (clip.transcriptionState) {
    return clip.transcriptionState;
  }
  
  // Fallback to legacy logic for backward compatibility
  if (clip.transcription && clip.transcription.length > 0) {
    return TranscriptionState.COMPLETED;
  }
  
  return TranscriptionState.IDLE;
}

/**
 * Check if a clip can be transcribed
 * @param {Object} clip - The video clip object
 * @returns {boolean} Whether the clip can be transcribed
 */
export function canTranscribe(clip) {
  const state = getTranscriptionState(clip);
  return state === TranscriptionState.IDLE || state === TranscriptionState.ERROR;
}

/**
 * Get a human-readable label for the transcription button
 * @param {Object} clip - The video clip object
 * @returns {string} The button label
 */
export function getTranscriptionButtonLabel(clip) {
  const state = getTranscriptionState(clip);
  
  switch (state) {
    case TranscriptionState.CHECKING:
      return 'Checking...';
    case TranscriptionState.TRANSCRIBING:
      return 'Transcribing...';
    case TranscriptionState.COMPLETED:
      return 'Transcribed';
    case TranscriptionState.ERROR:
      return 'Retry Transcription';
    default:
      return 'Transcribe';
  }
}

/**
 * Check if transcription is in progress
 * @param {Object} clip - The video clip object
 * @returns {boolean} Whether transcription is in progress
 */
export function isTranscribing(clip) {
  const state = getTranscriptionState(clip);
  return state === TranscriptionState.CHECKING || state === TranscriptionState.TRANSCRIBING;
}