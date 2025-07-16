<script lang="ts">
	import { onMount } from 'svelte';
	import anime from 'animejs/lib/anime.es.js';

	let { maxSnippets = 3 }: { maxSnippets?: number } = $props();

	let container: HTMLDivElement;
	let cleanupFunction: (() => void) | undefined;

	const codeSnippets = [
		// Single-line snippets
		{ text: 'build_together(your_vision)', type: 'function', multiline: false },
		{ text: 'solution = listen() + understand()', type: 'variable', multiline: false },
		{ text: 'deploy(simple, tested, reliable)', type: 'function', multiline: false },
		{ text: 'Solution.refactor(over_complicated)', type: 'method', multiline: false },
		{ text: 'Product(your_vision, our_expertise)', type: 'constructor', multiline: false },
		{ text: 'if needed: return honest_feedback()', type: 'condition', multiline: false },

		// Multi-line snippets
		{ text: 'if wrong_direction:\n    return redirect()', type: 'condition', multiline: true },
		{
			text: 'if too_complex:\n    simplify()\nelse:\n    test_and_ship()',
			type: 'condition',
			multiline: true
		},
		{
			text: 'try:\n    return build_perfect_solution()\nexcept OverEngineered:\n    return simplify()',
			type: 'exception',
			multiline: true
		},
		{
			text: 'def collaborate(your_needs):\n    """We listen first, build second"""\n    return solution',
			type: 'function',
			multiline: true
		},
		{ text: 'while business.grows():\n    we.adapt()', type: 'loop', multiline: true },
		{
			text: '# honest feedback saves money\nstart_simple(complex_feature)',
			type: 'comment',
			multiline: true
		}
	];

	onMount(() => {
		const initCodeAnimation = () => {
			if (!container) return;

			// Define mobile state and performance settings
			const isMobile = window.innerWidth < 1024;
			const isLowPowerDevice = isMobile && (
				navigator.hardwareConcurrency <= 4 || 
				(navigator as any).deviceMemory <= 4 ||
				/Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
			);
			
			// Reduce animations on mobile for better performance
			const enableFloatingAnimation = !isLowPowerDevice;
			const enableComplexAnimations = !isMobile;

			// Track active snippets to prevent overlap and limit count
			const activePositions = new Map<HTMLElement, {x: number, y: number}>();
			const activeSnippetElements = new Set<HTMLElement>();
			const usedSnippets = new Set<string>();
			const activeMultilineSnippets = new Set<string>();

			// Create syntax highlighted snippet
			const createHighlightedText = (snippet: { text: string; type: string }) => {
				const { text, type } = snippet;
				let html = text;

				// Python syntax highlighting
				if (type === 'function') {
					html = text.replace(/(\w+)\(/, '<span style="color: #7dd3fc">$1</span>(');
					html = html.replace(/(def)\s+/, '<span style="color: #f472b6">$1</span> ');
					html = html.replace(/(return)(\s)/g, '<span style="color: #f472b6">$1</span>$2');
				} else if (type === 'condition') {
					html = text.replace(
						/(if|else|return)(\s|:)/g,
						'<span style="color: #f472b6">$1</span>$2'
					);
				} else if (type === 'loop') {
					html = text.replace(/(while)(\s|:)/g, '<span style="color: #f472b6">$1</span>$2');
				} else if (type === 'variable') {
					html = text.replace(/(\w+)\s*=/, '<span style="color: #c084fc">$1</span> =');
				} else if (type === 'method') {
					html = text.replace(/\.(\w+)/, '.<span style="color: #7dd3fc">$1</span>');
				} else if (type === 'constructor') {
					html = text.replace(/(\w+)\(/, '<span style="color: #10b981">$1</span>(');
				} else if (type === 'exception') {
					html = text.replace(
						/(try|except|return)(\s|:)/g,
						'<span style="color: #f472b6">$1</span>$2'
					);
					html = html.replace(/(\w+Error)/g, '<span style="color: #ef4444">$1</span>');
				} else if (type === 'comment') {
					const lines = text.split('\n');
					const highlightedLines = lines.map((line) => {
						if (line.trim().startsWith('#')) {
							return `<span style="color: #86efac">${line}</span>`;
						} else {
							return line.replace(/(\w+)\(/g, '<span style="color: #7dd3fc">$1</span>(');
						}
					});
					html = highlightedLines.join('\n');
				}

				return html;
			};

			const measureTextDimensions = (text: string, fontSize: number, isMultiline: boolean) => {
				// Create temporary element to measure text
				const temp = document.createElement('div');
				temp.style.cssText = `
					position: absolute;
					visibility: hidden;
					white-space: ${isMultiline ? 'pre' : 'nowrap'};
					font-family: 'JetBrains Mono', 'Fira Code', 'Monaco', monospace;
					font-size: ${fontSize}px;
					line-height: ${isMultiline ? '1.3' : 'normal'};
					padding: 0;
					margin: 0;
				`;
				temp.innerHTML = text;
				document.body.appendChild(temp);
				
				const width = temp.offsetWidth;
				const height = temp.offsetHeight;
				
				document.body.removeChild(temp);
				return { width, height };
			};

			const getCenterPosition = () => {
				// Get container dimensions
				const containerRect = container.getBoundingClientRect();
				const containerWidth = containerRect.width;
				const containerHeight = containerRect.height;
				
				// Return center position
				return { 
					x: containerWidth / 2, 
					y: containerHeight / 2 
				};
			};

			const createSnippet = () => {
				// Don't create new snippet if we're at the limit
				if (activeSnippetElements.size >= maxSnippets) {
					return;
				}

				// Separate multi-line and single-line snippets
				const multilineSnippets = codeSnippets.filter((s) => s.multiline);
				const singlelineSnippets = codeSnippets.filter((s) => !s.multiline);

				// Check current multi-line count
				const currentMultilineCount = activeMultilineSnippets.size;

				let snippetData;
				let attempts = 0;

				// Force multi-line if we have none, or force single-line if we have 2
				if (currentMultilineCount === 0) {
					do {
						snippetData = multilineSnippets[Math.floor(Math.random() * multilineSnippets.length)];
						attempts++;
					} while (usedSnippets.has(snippetData.text) && attempts < 10);
				} else if (currentMultilineCount >= 2) {
					do {
						snippetData = singlelineSnippets[Math.floor(Math.random() * singlelineSnippets.length)];
						attempts++;
					} while (usedSnippets.has(snippetData.text) && attempts < 10);
				} else {
					// Can choose either, but prefer variety
					const availableSnippets = [...multilineSnippets, ...singlelineSnippets];
					do {
						snippetData = availableSnippets[Math.floor(Math.random() * availableSnippets.length)];
						attempts++;
					} while (usedSnippets.has(snippetData.text) && attempts < 10);
				}

				// If we can't find a unique snippet after 10 attempts, just use any snippet
				if (attempts >= 10) {
					snippetData = codeSnippets[Math.floor(Math.random() * codeSnippets.length)];
				}

				// Mark this snippet as in use
				usedSnippets.add(snippetData.text);

				// Track multi-line snippets
				if (snippetData.multiline) {
					activeMultilineSnippets.add(snippetData.text);
				}

				const snippet = document.createElement('div');
				snippet.className = 'code-snippet';
				snippet.innerHTML = createHighlightedText(snippetData);

				// Calculate font size first
				const depth = Math.random();
				const fontSize = isMobile ? 14 + depth * 4 : 16 + depth * 6;

				const position = getCenterPosition();
				activePositions.set(snippet, position);
				activeSnippetElements.add(snippet);

				// Depth-based properties for visual variety  
				const scale = 0.7 + depth * 0.3;
				const opacity = isMobile ? 0.2 + depth * 0.25 : 0.4 + depth * 0.35;
				const blur = depth > 0.7 ? 0.5 : 0;

				// Handle multi-line vs single-line snippets
				const whiteSpace = snippetData.multiline ? 'pre' : 'nowrap';
				const lineHeight = snippetData.multiline ? '1.3' : 'normal';

				// Generate unique animation names
				const animId = Math.random().toString(36).substr(2, 9);
				const fadeInClass = `fade-in-${animId}`;
				const floatClass = `float-${animId}`;
				
				snippet.style.cssText = `
					position: absolute;
					left: 50%;
					top: 50%;
					transform: translate(-50%, -50%);
					font-family: 'JetBrains Mono', 'Fira Code', 'Monaco', monospace;
					font-size: ${fontSize}px;
					color: rgb(156, 163, 175);
					white-space: ${whiteSpace};
					line-height: ${lineHeight};
					pointer-events: none;
					z-index: 1000;
					filter: blur(${blur}px);
					text-shadow: ${isMobile ? 'none' : '0 0 15px rgba(156, 163, 175, 0.4)'};
					opacity: 0;
					animation: ${fadeInClass} 0.8s ease-out forwards, ${floatClass} 8s ease-in-out infinite 0.8s;
				`;

				// Create CSS animations dynamically
				const styleSheet = document.createElement('style');
				styleSheet.textContent = `
					@keyframes ${fadeInClass} {
						from { opacity: 0; transform: translate(-50%, -50%) scale(${scale * 0.8}) translateZ(0); }
						to { opacity: ${opacity}; transform: translate(-50%, -50%) scale(${scale}) translateZ(0); }
					}
					@keyframes ${floatClass} {
						0%, 100% { transform: translate(-50%, -50%) scale(${scale}) translate(0px, 0px) translateZ(0); }
						25% { transform: translate(-50%, -50%) scale(${scale}) translate(4px, -3px) translateZ(0); }
						50% { transform: translate(-50%, -50%) scale(${scale}) translate(-4px, 3px) translateZ(0); }
						75% { transform: translate(-50%, -50%) scale(${scale}) translate(2px, -1px) translateZ(0); }
					}
				`;
				document.head.appendChild(styleSheet);

				container.appendChild(snippet);

				// Duration for snippet to stay on screen
				const displayDuration = 10000;
				const fadeOutDuration = 1000;
				
				// Schedule fade out and cleanup
				setTimeout(() => {
					// Add fade out animation
					const fadeOutClass = `fade-out-${animId}`;
					const fadeOutStyleSheet = document.createElement('style');
					fadeOutStyleSheet.textContent = `
						@keyframes ${fadeOutClass} {
							from { opacity: ${opacity}; }
							to { opacity: 0; }
						}
					`;
					document.head.appendChild(fadeOutStyleSheet);
					
					snippet.style.animation += `, ${fadeOutClass} ${fadeOutDuration}ms ease-out forwards`;
					
					// Remove after fade out
					setTimeout(() => {
						snippet.remove();
						activePositions.delete(snippet);
						activeSnippetElements.delete(snippet);
						usedSnippets.delete(snippetData.text);
						if (snippetData.multiline) {
							activeMultilineSnippets.delete(snippetData.text);
						}
						// Clean up style sheets
						styleSheet.remove();
						fadeOutStyleSheet.remove();
					}, fadeOutDuration);
				}, displayDuration);
			};

			// Staggered replacement system - optimized for device performance
			const snippetLifespan = isLowPowerDevice ? 15000 : isMobile ? 12000 : 10000;
			const staggerDelay = snippetLifespan / maxSnippets;
			let schedulerIntervalId: number;
			let nextSnippetIndex = 0;

			// Track snippet creation times for staggered replacement
			const snippetCreationTimes: number[] = [];

			const createStaggeredSnippet = () => {
				// Only create if we haven't reached the limit
				if (activeSnippetElements.size < maxSnippets) {
					createSnippet();
					snippetCreationTimes.push(Date.now());
				}
			};

			const maintainStaggeredFlow = () => {
				const now = Date.now();
				
				// Check if any snippets should be replaced
				if (snippetCreationTimes.length > 0) {
					const oldestCreationTime = snippetCreationTimes[0];
					const timeAlive = now - oldestCreationTime;
					
					// If the oldest snippet is about to fade out, create a new one
					if (timeAlive >= snippetLifespan - 2000 && activeSnippetElements.size >= maxSnippets) {
						createSnippet();
						snippetCreationTimes.shift(); // Remove the oldest timestamp
						snippetCreationTimes.push(now); // Add new timestamp
					}
				}
				
				// Ensure we maintain the target number of snippets
				if (activeSnippetElements.size < maxSnippets) {
					createSnippet();
					snippetCreationTimes.push(now);
				}
				
				// Continue the cycle
				schedulerIntervalId = setTimeout(maintainStaggeredFlow, 1000); // Check every second
			};

			// Start with initial snippets staggered over time
			for (let i = 0; i < maxSnippets; i++) {
				setTimeout(() => {
					createStaggeredSnippet();
				}, i * staggerDelay);
			}

			// Start the continuous staggered replacement system
			setTimeout(() => {
				maintainStaggeredFlow();
			}, maxSnippets * staggerDelay + 1000);

			// Cleanup function
			return () => {
				if (schedulerIntervalId) clearTimeout(schedulerIntervalId);
				container.querySelectorAll('.code-snippet').forEach(snippet => snippet.remove());
			};
		};

		// Initialize code animation
		setTimeout(() => {
			cleanupFunction = initCodeAnimation();
		}, 500);

		// Cleanup on destroy
		return () => {
			if (cleanupFunction) cleanupFunction();
		};
	});
</script>

<div
	bind:this={container}
	class="code-animation-container relative w-full h-full overflow-hidden pointer-events-none"
></div>
