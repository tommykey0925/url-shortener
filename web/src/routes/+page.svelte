<script lang="ts">
	import { shortenUrl, listUrls, deleteUrl, type URL as ShortURL, type ShortenResponse } from '$lib/api';

	let urls = $state<ShortURL[]>([]);
	let newUrl = $state('');
	let result = $state<ShortenResponse | null>(null);
	let error = $state('');
	let loading = $state(false);
	let copied = $state(false);

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
		copied = false;

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
		copied = true;
		setTimeout(() => (copied = false), 2000);
	}

	$effect(() => {
		loadUrls();
	});
</script>

<!-- Hero -->
<div class="mb-8 text-center sm:text-left">
	<p class="text-zinc-400">長いURLを短くして、クリック数も計測できます</p>
</div>

<!-- Shorten Form -->
<section class="mb-8">
	<form onsubmit={handleSubmit} class="flex flex-col gap-3 sm:flex-row">
		<input
			type="url"
			bind:value={newUrl}
			placeholder="https://example.com/very/long/url"
			required
			class="w-full rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-3 text-zinc-50 placeholder-zinc-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 sm:flex-1"
		/>
		<button
			type="submit"
			disabled={loading}
			class="rounded-lg bg-emerald-600 px-6 py-3 font-medium text-white transition hover:bg-emerald-500 disabled:opacity-50 sm:w-auto"
		>
			{loading ? '作成中...' : '短縮する'}
		</button>
	</form>

	{#if error}
		<div class="mt-3 rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-400">
			{error}
		</div>
	{/if}

	{#if result}
		<div class="mt-4 rounded-lg border border-emerald-800 bg-emerald-950 p-4">
			<p class="mb-2 text-xs text-emerald-400">短縮URLが作成されました</p>
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
				<a
					href={result.short_url}
					target="_blank"
					rel="noopener"
					class="break-all font-mono text-sm text-emerald-300 underline decoration-emerald-700 hover:decoration-emerald-400 sm:flex-1"
				>
					{result.short_url}
				</a>
				<button
					onclick={() => copyToClipboard(result!.short_url)}
					class="w-full rounded-md border border-emerald-700 px-4 py-2 text-sm text-emerald-300 transition hover:bg-emerald-900 sm:w-auto"
				>
					{copied ? 'コピーしました!' : 'コピー'}
				</button>
			</div>
		</div>
	{/if}
</section>

<!-- URL List -->
<section>
	<h2 class="mb-4 text-lg font-semibold">作成したURL</h2>

	{#if urls.length === 0}
		<div class="rounded-lg border border-dashed border-zinc-700 p-8 text-center">
			<p class="text-zinc-500">まだURLがありません。上のフォームから作成してみましょう</p>
		</div>
	{:else}
		<!-- Desktop: Table -->
		<div class="hidden overflow-hidden rounded-lg border border-zinc-800 sm:block">
			<table class="w-full text-sm">
				<thead class="border-b border-zinc-800 bg-zinc-900/50">
					<tr>
						<th class="px-4 py-3 text-left font-medium text-zinc-400">コード</th>
						<th class="px-4 py-3 text-left font-medium text-zinc-400">元のURL</th>
						<th class="px-4 py-3 text-right font-medium text-zinc-400">クリック数</th>
						<th class="px-4 py-3 text-right font-medium text-zinc-400">作成日</th>
						<th class="px-4 py-3"></th>
					</tr>
				</thead>
				<tbody>
					{#each urls as url (url.code)}
						<tr class="border-b border-zinc-800 transition last:border-0 hover:bg-zinc-900/30">
							<td class="px-4 py-3">
								<a
									href="/r/{url.code}"
									target="_blank"
									rel="noopener"
									class="font-mono text-emerald-400 hover:underline"
								>
									{url.code}
								</a>
							</td>
							<td class="max-w-xs truncate px-4 py-3 text-zinc-300" title={url.original_url}>
								{url.original_url}
							</td>
							<td class="px-4 py-3 text-right">
								<span class="inline-flex items-center rounded-full bg-zinc-800 px-2.5 py-0.5 text-xs font-medium text-zinc-300">
									{url.clicks}
								</span>
							</td>
							<td class="px-4 py-3 text-right text-zinc-500">
								{new Date(url.created_at).toLocaleDateString('ja-JP')}
							</td>
							<td class="px-4 py-3 text-right">
								<button
									onclick={() => handleDelete(url.code)}
									class="rounded px-2 py-1 text-xs text-zinc-500 transition hover:bg-red-950 hover:text-red-400"
								>
									削除
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Mobile: Card List -->
		<div class="flex flex-col gap-3 sm:hidden">
			{#each urls as url (url.code)}
				<div class="rounded-lg border border-zinc-800 bg-zinc-900/30 p-4">
					<div class="mb-2 flex items-center justify-between">
						<a
							href="/r/{url.code}"
							target="_blank"
							rel="noopener"
							class="font-mono text-emerald-400 hover:underline"
						>
							{url.code}
						</a>
						<span class="inline-flex items-center rounded-full bg-zinc-800 px-2.5 py-0.5 text-xs font-medium text-zinc-300">
							{url.clicks} clicks
						</span>
					</div>
					<p class="mb-3 truncate text-sm text-zinc-400" title={url.original_url}>
						{url.original_url}
					</p>
					<div class="flex items-center justify-between">
						<span class="text-xs text-zinc-600">
							{new Date(url.created_at).toLocaleDateString('ja-JP')}
						</span>
						<button
							onclick={() => handleDelete(url.code)}
							class="rounded px-3 py-1 text-xs text-zinc-500 transition hover:bg-red-950 hover:text-red-400"
						>
							削除
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</section>
