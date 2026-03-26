<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>URL Shortener</title>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</svelte:head>

<div class="min-h-screen bg-zinc-950 text-zinc-50">
	<header class="border-b border-zinc-800 px-4 py-4 sm:px-6">
		<div class="mx-auto max-w-3xl">
			<h1 class="text-xl font-bold tracking-tight">
				<span class="text-emerald-400">/</span> URL Shortener
			</h1>
		</div>
	</header>
	<main class="mx-auto max-w-3xl px-4 py-6 sm:px-6 sm:py-8">
		{@render children()}
	</main>
	<footer class="border-t border-zinc-900 px-4 py-4 text-center text-xs text-zinc-600">
		Go + SvelteKit + EKS + ArgoCD + Terraform
	</footer>
</div>
