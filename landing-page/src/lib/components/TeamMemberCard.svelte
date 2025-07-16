<script lang="ts">
	let {
		member,
		index = 0
	}: {
		member: {
			name: string;
			role: string;
			bio: string;
			image: string;
			skills: string[];
		};
		index?: number;
	} = $props();

	const colorClass = $derived(index % 3 === 0 ? 'green' : index % 3 === 1 ? 'pink' : 'cyan');
</script>

<div class="group text-center space-y-4">
	<!-- Profile Picture -->
	<div class="relative mx-auto w-fit">
		<img
			src={member.image}
			alt={member.name}
			loading="lazy"
			decoding="async"
			class="w-24 h-24 rounded-full object-cover ring-border group-hover:ring-{colorClass}-500/50 transition-all duration-300 group-hover:scale-105"
		/>
	</div>

	<!-- Member Info -->
	<div class="space-y-2">
		<h3 class="text-lg font-bold text-foreground">{member.name}</h3>
		<p class="text-{colorClass}-400 font-medium text-sm">
			{member.role}
		</p>
		<p class="text-muted-foreground text-sm leading-relaxed max-w-xs mx-auto">{member.bio}</p>
	</div>

</div>