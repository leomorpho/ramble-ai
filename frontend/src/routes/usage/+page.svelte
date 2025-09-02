<!-- 
🔗 CROSS-REFERENCE: This component mirrors the PocketBase web frontend
📁 Location: pb-be/sk/src/routes/(protected)/usages/+page.svelte  
⚠️  Any changes to this usage statistics component should be reflected in the PocketBase version
-->

<script>
  import { onMount } from "svelte";
  import { GetRambleAIApiKey, GetBackendURL } from "$lib/wailsjs/go/main/App";
  import { Button } from "$lib/components/ui/button";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import { 
    BarChart, 
    FileAudio, 
    Clock, 
    TrendingUp, 
    Download, 
    Calendar,
    CheckCircle2,
    AlertCircle,
    Loader2,
    ArrowLeft,
    RefreshCw
  } from "@lucide/svelte";

  // TypeScript interfaces matching PocketBase frontend
  /** @typedef {Object} ProcessedFile
   * @property {string} id
   * @property {string} filename
   * @property {number} file_size_bytes
   * @property {number} duration_seconds
   * @property {number} processing_time_ms
   * @property {'completed'|'processing'|'failed'} status
   * @property {number} transcript_length
   * @property {number} words_count
   * @property {string} model_used
   * @property {string} created
   * @property {string} updated
   */

  /** @typedef {Object} UsageSummary
   * @property {number} total_files
   * @property {number} total_duration_seconds
   * @property {number} total_duration_minutes
   * @property {number} total_duration_hours
   * @property {number} total_file_size_bytes
   * @property {number} total_file_size_mb
   * @property {number} total_processing_time_ms
   * @property {number} avg_processing_time_ms
   * @property {Object} status_breakdown
   * @property {number} status_breakdown.completed
   * @property {number} status_breakdown.processing
   * @property {number} status_breakdown.failed
   * @property {number} success_rate
   */

  /** @type {ProcessedFile[]} */
  let processedFiles = $state([]);
  /** @type {UsageSummary|null} */
  let currentMonthSummary = $state(null);
  /** @type {UsageSummary|null} */
  let allTimeSummary = $state(null);
  let isLoading = $state(true);
  /** @type {string|null} */
  let error = $state(null);
  let selectedMonth = $state(new Date().toISOString().slice(0, 7)); // YYYY-MM

  // PocketBase backend base URL - get from Go backend
  let POCKETBASE_URL = $state("https://api.ramble.goosebyteshq.com"); // fallback

  onMount(async () => {
    // Get the backend URL from Go configuration
    try {
      POCKETBASE_URL = await GetBackendURL();
      console.log(`Using PocketBase URL: ${POCKETBASE_URL}`);
    } catch (err) {
      console.warn('Failed to get backend URL, using fallback:', err);
    }
    
    await loadUsageData();
    
    // Add page visibility listener to refresh data when user returns to page
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        console.log('Page became visible, refreshing usage data...');
        loadUsageData();
      }
    };
    
    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    // Cleanup on unmount
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  });

  async function loadUsageData() {
    try {
      isLoading = true;
      error = null;

      // Get Ramble AI API key for authentication
      const apiKey = await GetRambleAIApiKey();
      if (!apiKey) {
        error = "Ramble AI API key not configured. Please set it up in Settings.";
        return;
      }

      // Load processed files (recent ones for the table)
      const filesResponse = await fetch(`${POCKETBASE_URL}/api/usage/files?page=1&perPage=50`, {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json'
        }
      });

      if (!filesResponse.ok) {
        throw new Error(`Failed to load files: ${filesResponse.status} ${filesResponse.statusText}`);
      }

      const filesData = await filesResponse.json();
      console.log('Files API response:', filesData); // Debug logging
      processedFiles = filesData.files || [];

      // Load usage summary
      const summaryResponse = await fetch(`${POCKETBASE_URL}/api/usage/summary`, {
        headers: {
          'Authorization': `Bearer ${apiKey}`,
          'Content-Type': 'application/json'
        }
      });

      if (!summaryResponse.ok) {
        throw new Error(`Failed to load summary: ${summaryResponse.status} ${summaryResponse.statusText}`);
      }

      const summaryData = await summaryResponse.json();
      console.log('Summary API response:', summaryData); // Debug logging
      
      // For now, use all-time data for both summaries until we add date filtering
      allTimeSummary = summaryData;
      currentMonthSummary = summaryData;

    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load usage data';
      console.error('Error loading usage data:', err);
    } finally {
      isLoading = false;
    }
  }

  function calculateSummary(files) {
    const totalFiles = files.length;
    let totalDuration = 0;
    let totalFileSize = 0;
    let totalProcessingTime = 0;
    const statusCounts = { completed: 0, processing: 0, failed: 0 };

    for (const file of files) {
      totalDuration += file.duration_seconds || 0;
      totalFileSize += file.file_size_bytes || 0;
      totalProcessingTime += file.processing_time_ms || 0;
      
      if (statusCounts.hasOwnProperty(file.status)) {
        statusCounts[file.status]++;
      }
    }

    const totalMinutes = totalDuration / 60;
    const totalHours = totalMinutes / 60;
    const avgProcessingTime = totalFiles > 0 ? totalProcessingTime / totalFiles : 0;
    const successRate = totalFiles > 0 ? (statusCounts.completed / totalFiles) * 100 : 0;

    return {
      total_files: totalFiles,
      total_duration_seconds: totalDuration,
      total_duration_minutes: totalMinutes,
      total_duration_hours: totalHours,
      total_file_size_bytes: totalFileSize,
      total_file_size_mb: totalFileSize / (1024 * 1024),
      total_processing_time_ms: totalProcessingTime,
      avg_processing_time_ms: avgProcessingTime,
      status_breakdown: statusCounts,
      success_rate: successRate,
    };
  }

  function formatDuration(seconds) {
    if (seconds < 60) {
      return `${Math.round(seconds)}s`;
    } else if (seconds < 3600) {
      return `${Math.round(seconds / 60)}m`;
    } else {
      const hours = Math.floor(seconds / 3600);
      const minutes = Math.round((seconds % 3600) / 60);
      return `${hours}h ${minutes}m`;
    }
  }

  function formatFileSize(bytes) {
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)}MB`;
    return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)}GB`;
  }

  function formatDate(dateString) {
    if (!dateString) return 'Unknown';
    
    const date = new Date(dateString);
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      console.warn('Invalid date received:', dateString);
      return 'Invalid Date';
    }
    
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  function getStatusIcon(status) {
    switch (status) {
      case 'completed': return CheckCircle2;
      case 'processing': return Loader2;
      case 'failed': return AlertCircle;
      default: return AlertCircle;
    }
  }

  function getStatusColor(status) {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
      case 'processing': return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
      case 'failed': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    }
  }

  function hasBeenMoreThanOneMonth() {
    // For Wails app, we need to get user data from the backend
    // For now, we'll return true to maintain existing behavior until we implement proper user auth
    return true; // TODO: Implement proper user registration date check when Wails auth is implemented
  }
</script>

<main class="min-h-screen bg-background text-foreground p-8">
  <div class="max-w-6xl mx-auto space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div class="flex items-center gap-3">
        <a href="/" class="text-muted-foreground hover:text-foreground" aria-label="Back to home">
          <ArrowLeft class="w-4 h-4" />
        </a>
        <BarChart class="h-8 w-8 text-primary" />
        <h1 class="text-3xl font-bold text-foreground">Usage Statistics</h1>
      </div>
      <Button onclick={loadUsageData} variant="outline" size="sm" disabled={isLoading}>
        <RefreshCw class={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
        Refresh
      </Button>
    </div>
    <p class="text-muted-foreground -mt-6 ml-11">Track your video processing usage and history</p>
    
    <div class="mt-4 ml-11 p-3 bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg">
      <p class="text-sm text-green-800 dark:text-green-200">
        🔒 <strong>Privacy First:</strong> All audio and video processing happens locally on your machine. 
        We never store your files on our servers - only processing metadata is tracked for usage statistics.
      </p>
    </div>

    {#if isLoading}
      <!-- Skeleton Loading State -->
      <div class="space-y-6">
        <!-- Summary Cards Skeleton -->
        <div class="grid gap-6 md:grid-cols-2 mb-8">
          <!-- Card 1 -->
          <div class="border rounded p-6 bg-card">
            <div class="flex items-center justify-between mb-2">
              <Skeleton class="h-4 w-20" />
              <Skeleton class="h-4 w-4" />
            </div>
            <div class="space-y-2">
              <Skeleton class="h-8 w-16" />
              <Skeleton class="h-4 w-24" />
            </div>
          </div>
          
          <!-- Card 2 -->
          <div class="border rounded p-6 bg-card">
            <div class="flex items-center justify-between mb-2">
              <Skeleton class="h-4 w-24" />
              <Skeleton class="h-4 w-4" />
            </div>
            <div class="space-y-2">
              <Skeleton class="h-8 w-12" />
              <Skeleton class="h-4 w-20" />
            </div>
          </div>
          
        </div>

        <!-- All-Time Statistics Skeleton -->
        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {#each Array(4) as _, i}
            <div class="border rounded p-6 bg-card">
              <div class="flex items-center justify-between mb-2">
                <Skeleton class="h-4 w-20" />
                <Skeleton class="h-4 w-4" />
              </div>
              <div class="space-y-2">
                <Skeleton class="h-8 w-16" />
                <Skeleton class="h-4 w-24" />
              </div>
            </div>
          {/each}
        </div>

        <!-- Recent Files Table Skeleton -->
        <div class="border rounded bg-card">
          <div class="border-b border-border p-4">
            <Skeleton class="h-6 w-32" />
          </div>
          <div class="overflow-x-auto">
            <table class="w-full">
              <thead class="bg-muted/50">
                <tr class="border-b border-border/50">
                  <th class="text-left py-3 px-4">
                    <Skeleton class="h-4 w-16" />
                  </th>
                  <th class="text-left py-3 px-4">
                    <Skeleton class="h-4 w-12" />
                  </th>
                  <th class="text-left py-3 px-4">
                    <Skeleton class="h-4 w-20" />
                  </th>
                  <th class="text-left py-3 px-4">
                    <Skeleton class="h-4 w-16" />
                  </th>
                  <th class="text-left py-3 px-4">
                    <Skeleton class="h-4 w-12" />
                  </th>
                  <th class="text-left py-3 px-4">
                    <Skeleton class="h-4 w-20" />
                  </th>
                </tr>
              </thead>
              <tbody>
                {#each Array(5) as _, i}
                  <tr class="border-b border-border/50">
                    <td class="py-4 px-4">
                      <div class="flex items-center gap-3">
                        <Skeleton class="h-4 w-4" />
                        <div class="space-y-1">
                          <Skeleton class="h-4 w-32" />
                          <Skeleton class="h-3 w-20" />
                        </div>
                      </div>
                    </td>
                    <td class="py-4 px-4">
                      <Skeleton class="h-5 w-16 rounded-full" />
                    </td>
                    <td class="py-4 px-4">
                      <Skeleton class="h-4 w-12" />
                    </td>
                    <td class="py-4 px-4">
                      <Skeleton class="h-4 w-16" />
                    </td>
                    <td class="py-4 px-4">
                      <Skeleton class="h-4 w-20" />
                    </td>
                    <td class="py-4 px-4">
                      <Skeleton class="h-4 w-24" />
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    {:else if error}
      <div class="rounded-md border border-red-200 bg-red-50 p-4 text-red-600 dark:border-red-900 dark:bg-red-950 dark:text-red-400">
        <div class="flex items-center gap-2">
          <AlertCircle class="h-5 w-5" />
          <span>Error loading usage data: {error}</span>
        </div>
        <Button onclick={loadUsageData} variant="outline" class="mt-3">
          Retry
        </Button>
      </div>
    {:else}
      <!-- Summary Cards -->
      <div class="grid gap-6 {hasBeenMoreThanOneMonth() ? 'md:grid-cols-2' : 'md:grid-cols-1'} mb-8">
        <!-- Current Month Summary -->
        {#if hasBeenMoreThanOneMonth() && currentMonthSummary}
          <div class="border rounded p-6 bg-card">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-muted-foreground">This Month</span>
              <Calendar class="h-4 w-4 text-muted-foreground" />
            </div>
            <div class="space-y-2">
              <div class="text-2xl font-bold">{formatDuration(currentMonthSummary.total_duration_seconds)}</div>
              <div class="text-sm text-muted-foreground">
                {currentMonthSummary.total_files} files processed
              </div>
            </div>
          </div>
        {/if}

        <!-- All Time Summary -->
        {#if allTimeSummary}
          <div class="border rounded p-6 bg-card">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-muted-foreground">Total Processing</span>
              <Clock class="h-4 w-4 text-muted-foreground" />
            </div>
            <div class="space-y-2">
              <div class="text-2xl font-bold">{formatDuration(allTimeSummary.total_duration_seconds)}</div>
              <div class="text-sm text-muted-foreground">
                {allTimeSummary.total_files} files all-time
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Recent Files Table -->
      <div class="border rounded overflow-hidden bg-card">
        <div class="p-6 pb-4">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-xl font-semibold">Recent Processing History</h2>
            <Button variant="outline" size="sm">
              <Download class="h-4 w-4 mr-2" />
              Export
            </Button>
          </div>
          
          {#if processedFiles.length === 0}
            <div class="text-center py-12 text-muted-foreground">
              <FileAudio class="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p class="text-lg font-medium">No files processed yet</p>
              <p class="text-sm">Your transcription history will appear here once you start processing audio files.</p>
            </div>
          {:else}
            <div class="overflow-x-auto">
              <table class="w-full">
                <thead>
                  <tr class="border-b border-border text-left">
                    <th class="pb-3 text-sm font-medium text-muted-foreground">File</th>
                    <th class="pb-3 text-sm font-medium text-muted-foreground">Duration</th>
                    <th class="pb-3 text-sm font-medium text-muted-foreground">Size</th>
                    <th class="pb-3 text-sm font-medium text-muted-foreground">Status</th>
                    <th class="pb-3 text-sm font-medium text-muted-foreground">Processed</th>
                  </tr>
                </thead>
                <tbody>
                  {#each processedFiles as file}
                    <tr class="border-b border-border/50">
                      <td class="py-4">
                        <div class="flex items-center gap-3">
                          <FileAudio class="h-4 w-4 text-muted-foreground" />
                          <div>
                            <div class="font-medium text-sm">{file.filename}</div>
                            <div class="text-xs text-muted-foreground">
                              {file.words_count} words, {file.transcript_length} chars
                            </div>
                          </div>
                        </div>
                      </td>
                      <td class="py-4 text-sm">
                        {formatDuration(file.duration_seconds)}
                      </td>
                      <td class="py-4 text-sm">
                        {formatFileSize(file.file_size_bytes)}
                      </td>
                      <td class="py-4">
                        {#each [file.status] as status}
                          {@const IconComponent = getStatusIcon(status)}
                          <div class={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border-0 ${getStatusColor(status)}`}>
                            <IconComponent class={`h-3 w-3 mr-1 ${status === 'processing' ? 'animate-spin' : ''}`} />
                            {status}
                          </div>
                        {/each}
                      </td>
                      <td class="py-4 text-sm text-muted-foreground">
                        {formatDate(file.created)}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>
      </div>
    {/if}

  </div>
</main>