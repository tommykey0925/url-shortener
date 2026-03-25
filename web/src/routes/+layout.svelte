<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>URL Shortener</title>
</svelte:head>

<div class="min-h-screen bg-zinc-950 text-zinc-50">
	<header class="border-b border-zinc-800 px-6 py-4">
		<h1 class="text-xl font-bold">URL Shortener</h1>
	</header>
	<main class="mx-auto max-w-3xl px-6 py-8">
		{@render children()}
	</main>
</div>
