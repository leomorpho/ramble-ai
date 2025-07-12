<script>
  import { AIChatbot } from "./index.js";
  import { CHATBOT_ENDPOINTS, CHATBOT_POSITIONS } from "$lib/constants/chatbot.js";
  
  // Example usage props
  let {
    projectId = 1,
    highlightData = [],
    orderData = []
  } = $props();
  
  // Example context data that would be sent to the AI
  $: contextData = {
    highlights: highlightData,
    order: orderData,
    projectInfo: {
      id: projectId,
      totalHighlights: highlightData.length
    }
  };
</script>

<!-- Example 1: Floating brain button (default) -->
<AIChatbot 
  endpointId={CHATBOT_ENDPOINTS.HIGHLIGHT_ORDERING}
  {projectId}
  {contextData}
  position={CHATBOT_POSITIONS.FLOATING}
/>

<!-- Example 2: Inline button (for replacing AISettings) -->
<AIChatbot 
  endpointId={CHATBOT_ENDPOINTS.HIGHLIGHT_SUGGESTIONS}
  {projectId}
  {contextData}
  position={CHATBOT_POSITIONS.INLINE}
  buttonText="Get AI Suggestions"
  className="w-full"
/>