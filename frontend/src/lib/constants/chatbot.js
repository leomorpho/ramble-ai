// Chatbot endpoint identifiers
export const CHATBOT_ENDPOINTS = {
  HIGHLIGHT_ORDERING: 'highlight_ordering',
  HIGHLIGHT_SUGGESTIONS: 'highlight_suggestions',
  CONTENT_ANALYSIS: 'content_analysis',
  EXPORT_OPTIMIZATION: 'export_optimization'
};

// Configuration for each chatbot endpoint
export const ENDPOINT_CONFIGS = {
  [CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING]: {
    title: 'Highlight Ordering Assistant',
    description: 'Help with organizing and reordering highlights for better flow',
    systemPrompt: 'You are an expert video content organizer. Help the user organize their video highlights for optimal storytelling and engagement.',
    icon: '🎬',
    defaultModel: 'anthropic/claude-sonnet-4'
  },
  [CHATBOT_ENDPOINTS.HIGHLIGHT_SUGGESTIONS]: {
    title: 'Highlight Suggestions Assistant',
    description: 'Get AI suggestions for creating engaging highlights',
    systemPrompt: 'You are an expert at identifying compelling moments in video content. Help suggest highlights that will engage viewers.',
    icon: '✨',
    defaultModel: 'anthropic/claude-sonnet-4'
  },
  [CHATBOT_ENDPOINTS.CONTENT_ANALYSIS]: {
    title: 'Content Analysis Assistant',
    description: 'Analyze video content for insights and recommendations',
    systemPrompt: 'You are a content analysis expert. Help analyze video content for themes, key messages, and audience engagement opportunities.',
    icon: '📊',
    defaultModel: 'google/gemini-2.0-flash-001'
  },
  [CHATBOT_ENDPOINTS.EXPORT_OPTIMIZATION]: {
    title: 'Export Optimization Assistant',
    description: 'Optimize export settings and final video production',
    systemPrompt: 'You are a video production expert. Help optimize export settings and final video production for different platforms and audiences.',
    icon: '🚀',
    defaultModel: 'anthropic/claude-3.5-haiku-20241022'
  }
};

// Available AI models
export const AVAILABLE_MODELS = [
  { value: "anthropic/claude-sonnet-4", label: "Claude Sonnet 4 (Latest)" },
  { value: "google/gemini-2.0-flash-001", label: "Gemini 2.0 Flash" },
  { value: "google/gemini-2.5-flash-preview-05-20", label: "Gemini 2.5 Flash Preview" },
  { value: "deepseek/deepseek-chat-v3-0324:free", label: "DeepSeek Chat v3 (Free)" },
  { value: "anthropic/claude-3.7-sonnet", label: "Claude 3.7 Sonnet" },
  { value: "anthropic/claude-3.5-haiku-20241022", label: "Claude 3.5 Haiku (Fast)" },
  { value: "openai/gpt-4o-mini", label: "GPT-4o Mini" },
  { value: "mistralai/mistral-nemo", label: "Mistral Nemo" },
  { value: "custom", label: "Custom Model" }
];

// Chat message types
export const MESSAGE_TYPES = {
  USER: 'user',
  ASSISTANT: 'assistant',
  SYSTEM: 'system',
  ERROR: 'error'
};

// Chatbot positioning options
export const CHATBOT_POSITIONS = {
  FLOATING: 'floating',
  INLINE: 'inline',
  SHEET: 'sheet'
};