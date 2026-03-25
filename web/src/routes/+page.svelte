<script lang="ts">
	import { shortenUrl, listUrls, deleteUrl, type URL as ShortURL, type ShortenResponse } from '$lib/api';

	let urls = $state<ShortURL[]>([]);
	let newUrl = $state('');
	let result = $state<ShortenResponse | null>(null);
	let error = $state('');
	let loading = $state(false);

	async function loadUrls() {
		try {
			urls = (await listUrls()) ?? [];
		} catch {
			urls = [];
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!newUrl.trim()) return;

		loading = true;
		error = '';
		result = null;

		try {
			result = await shortenUrl(newUrl);
			newUrl = '';
			await loadUrls();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Something went wrong';
		} finally {
			loading = false;
		}
	}

	async function handleDelete(code: string) {
		try {
			await deleteUrl(code);
			await loadUrls();
		} catch {
			error = 'Failed to delete URL';
		}
	}

	async function copyToClipboard(text: string) {
		await navigator.clipboard.writeText(text);
	}

	$effect(() => {
		loadUrls();
	});
</script>

<!-- Shorten Form -->
<section class="mb-8">
	<form onsubmit={handleSubmit} class="flex gap-3">
		<input
			type="url"
			bind:value={newUrl}
			placeholder="https://example.com/very/long/url"
			required
			class="flex-1 rounded-md border border-zinc-700 bg-zinc-900 px-4 py-2 text-zinc-50 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
		/>
		<button
			type="submit"
			disabled={loading}
			class="rounded-md bg-zinc-50 px-6 py-2 font-medium text-zinc-900 hover:bg-zinc-200 disabled:opacity-50"
		>
			{loading ? 'Shortening...' : 'Shorten'}
		</button>
	</form>

	{#if error}
		<p class="mt-3 text-sm text-red-400">{error}</p>
	{/if}

	{#if result}
		<div class="mt-4 flex items-center gap-3 rounded-md border border-zinc-700 bg-zinc-900 px-4 py-3">
			<span class="flex-1 font-mono text-sm text-emerald-400">{result.short_url}</span>
			<button
				onclick={() => copyToClipboard(result!.short_url)}
				class="rounded px-3 py-1 text-sm text-zinc-400 hover:bg-zinc-800 hover:text-zinc-50"
			>
				Copy
			</button>
		</div>
	{/if}
</section>

<!-- URL List -->
<section>
	<h2 class="mb-4 text-lg font-semibold">Your URLs</h2>

	{#if urls.length === 0}
		<p class="text-zinc-500">No URLs yet. Shorten one above!</p>
	{:else}
		<div class="overflow-hidden rounded-md border border-zinc-800">
			<table class="w-full text-sm">
				<thead class="border-b border-zinc-800 bg-zinc-900">
					<tr>
						<th class="px-4 py-3 text-left font-medium text-zinc-400">Code</th>
						<th class="px-4 py-3 text-left font-medium text-zinc-400">Original URL</th>
						<th class="px-4 py-3 text-right font-medium text-zinc-400">Clicks</th>
						<th class="px-4 py-3 text-right font-medium text-zinc-400">Created</th>
						<th class="px-4 py-3"></th>
					</tr>
				</thead>
				<tbody>
					{#each urls as url (url.code)}
						<tr class="border-b border-zinc-800 last:border-0">
							<td class="px-4 py-3 font-mono text-emerald-400">{url.code}</td>
							<td class="max-w-xs truncate px-4 py-3 text-zinc-300" title={url.original_url}>
								{url.original_url}
							</td>
							<td class="px-4 py-3 text-right text-zinc-300">{url.clicks}</td>
							<td class="px-4 py-3 text-right text-zinc-500">
								{new Date(url.created_at).toLocaleDateString()}
							</td>
							<td class="px-4 py-3 text-right">
								<button
									onclick={() => handleDelete(url.code)}
									class="text-zinc-500 hover:text-red-400"
								>
									Delete
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</section>
