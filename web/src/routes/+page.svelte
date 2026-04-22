<script lang="ts">
	import { shortenUrl, listUrls, deleteUrl, summarizeUrl, getClickStats, type URL as ShortURL, type ShortenResponse, type ClickStats } from '$lib/api';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Badge } from '$lib/components/ui/badge';
	import * as Alert from '$lib/components/ui/alert';

	let urls = $state<ShortURL[]>([]);
	let newUrl = $state('');
	let result = $state<ShortenResponse | null>(null);
	let error = $state('');
	let loading = $state(false);
	let copied = $state(false);
	let summarizing = $state('');
	let summaryText = $state('');
	let showSummary = $state(false);
	let statsLoading = $state('');
	let statsData = $state<ClickStats | null>(null);
	let showStats = $state(false);

	function getMyCodes(): string[] {
		try {
			const raw = localStorage.getItem('my_urls');
			return raw ? JSON.parse(raw) : [];
		} catch { return []; }
	}

	function addMyCode(code: string) {
		const codes = getMyCodes();
		codes.push(code);
		localStorage.setItem('my_urls', JSON.stringify(codes));
	}

	function removeMyCode(code: string) {
		const codes = getMyCodes().filter(c => c !== code);
		localStorage.setItem('my_urls', JSON.stringify(codes));
	}

	async function loadUrls() {
		try {
			const all = (await listUrls()) ?? [];
			const myCodes = getMyCodes();
			urls = all.filter(u => myCodes.includes(u.code));
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
			addMyCode(result.code);
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
			removeMyCode(code);
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

	async function handleSummarize(code: string) {
		summarizing = code;
		summaryText = '';
		try {
			summaryText = await summarizeUrl(code);
			showSummary = true;
		} catch {
			summaryText = '要約の取得に失敗しました';
			showSummary = true;
		} finally {
			summarizing = '';
		}
	}

	async function handleStats(code: string) {
		statsLoading = code;
		statsData = null;
		try {
			statsData = await getClickStats(code);
			showStats = true;
		} catch {
			error = 'Failed to load stats';
		} finally {
			statsLoading = '';
		}
	}

	$effect(() => {
		loadUrls();
	});
</script>

<!-- Hero -->
<div class="mb-8 text-center sm:text-left">
	<p class="text-muted-foreground">長いURLを短くして、クリック数も計測できます</p>
</div>

<!-- Shorten Form -->
<section class="mb-8">
	<form onsubmit={handleSubmit} class="flex flex-col gap-3 sm:flex-row">
		<Input
			type="url"
			bind:value={newUrl}
			placeholder="https://example.com/very/long/url"
			required
			class="sm:flex-1"
		/>
		<Button type="submit" disabled={loading}>
			{loading ? '作成中...' : '短縮する'}
		</Button>
	</form>

	{#if error}
		<Alert.Root variant="destructive" class="mt-3">
			<Alert.Description>{error}</Alert.Description>
		</Alert.Root>
	{/if}

	{#if result}
		<Card.Root class="mt-4 border-primary/30 bg-primary/5">
			<Card.Content class="pt-4">
				<p class="mb-2 text-xs text-primary">短縮URLが作成されました</p>
				<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
					<a
						href={result.short_url}
						target="_blank"
						rel="noopener"
						class="break-all font-mono text-sm text-primary underline decoration-primary/40 hover:decoration-primary sm:flex-1"
					>
						{result.short_url}
					</a>
					<Button variant="outline" size="sm" onclick={() => copyToClipboard(result!.short_url)}>
						{copied ? 'コピーしました!' : 'コピー'}
					</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</section>

<!-- URL List -->
<section>
	<h2 class="mb-4 text-lg font-semibold">作成したURL</h2>

	{#if urls.length === 0}
		<Card.Root class="border-dashed">
			<Card.Content class="py-8 text-center">
				<p class="text-muted-foreground">まだURLがありません。上のフォームから作成してみましょう</p>
			</Card.Content>
		</Card.Root>
	{:else}
		<!-- Desktop: Table -->
		<div class="hidden sm:block">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>コード</Table.Head>
						<Table.Head>元のURL</Table.Head>
						<Table.Head class="text-right">クリック数</Table.Head>
						<Table.Head class="text-right">作成日</Table.Head>
						<Table.Head></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each urls as url (url.code)}
						<Table.Row>
							<Table.Cell>
								<a
									href="/r/{url.code}"
									target="_blank"
									rel="noopener"
									class="font-mono text-primary hover:underline"
								>
									{url.code}
								</a>
							</Table.Cell>
							<Table.Cell class="max-w-xs truncate" title={url.original_url}>
								{url.original_url}
							</Table.Cell>
							<Table.Cell class="text-right">
								<Badge variant="secondary">{url.clicks}</Badge>
							</Table.Cell>
							<Table.Cell class="text-right text-muted-foreground">
								{new Date(url.created_at).toLocaleDateString('ja-JP')}
							</Table.Cell>
							<Table.Cell class="text-right">
								<div class="flex justify-end gap-1">
									<Button
										variant="ghost"
										size="sm"
										disabled={statsLoading === url.code}
										onclick={() => handleStats(url.code)}
									>
										{statsLoading === url.code ? '読込中...' : '統計'}
									</Button>
									<Button
										variant="ghost"
										size="sm"
										disabled={summarizing === url.code}
										onclick={() => handleSummarize(url.code)}
									>
										{summarizing === url.code ? '分析中...' : 'AI要約'}
									</Button>
									<Button
										variant="ghost"
										size="sm"
										class="text-destructive hover:text-destructive"
										onclick={() => handleDelete(url.code)}
									>
										削除
									</Button>
								</div>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		<!-- Mobile: Card List -->
		<div class="flex flex-col gap-3 sm:hidden">
			{#each urls as url (url.code)}
				<Card.Root>
					<Card.Content class="pt-4">
						<div class="mb-2 flex items-center justify-between">
							<a
								href="/r/{url.code}"
								target="_blank"
								rel="noopener"
								class="font-mono text-primary hover:underline"
							>
								{url.code}
							</a>
							<Badge variant="secondary">{url.clicks} clicks</Badge>
						</div>
						<p class="mb-3 truncate text-sm text-muted-foreground" title={url.original_url}>
							{url.original_url}
						</p>
						<div class="flex items-center justify-between">
							<span class="text-xs text-muted-foreground">
								{new Date(url.created_at).toLocaleDateString('ja-JP')}
							</span>
							<div class="flex gap-1">
								<Button
									variant="ghost"
									size="sm"
									disabled={statsLoading === url.code}
									onclick={() => handleStats(url.code)}
								>
									{statsLoading === url.code ? '読込中...' : '統計'}
								</Button>
								<Button
									variant="ghost"
									size="sm"
									disabled={summarizing === url.code}
									onclick={() => handleSummarize(url.code)}
								>
									{summarizing === url.code ? '分析中...' : 'AI要約'}
								</Button>
								<Button
									variant="ghost"
									size="sm"
									class="text-destructive hover:text-destructive"
									onclick={() => handleDelete(url.code)}
								>
									削除
								</Button>
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</section>

<!-- Click Stats Dialog -->
<Dialog.Root bind:open={showStats}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>クリック統計 — {statsData?.code}</Dialog.Title>
		</Dialog.Header>
		{#if statsData}
			<div class="space-y-4">
				<div class="text-center">
					<p class="text-3xl font-bold text-primary">{statsData.total_clicks}</p>
					<p class="text-sm text-muted-foreground">総クリック数</p>
				</div>
				{#if statsData.daily.length > 0}
					<div>
						<p class="mb-2 text-sm font-medium">日別クリック数（過去30日）</p>
						<div class="max-h-60 overflow-y-auto">
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>日付</Table.Head>
										<Table.Head class="text-right">クリック数</Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each statsData.daily as day}
										<Table.Row>
											<Table.Cell>{day.date}</Table.Cell>
											<Table.Cell class="text-right">
												<Badge variant="secondary">{day.clicks}</Badge>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						</div>
					</div>
				{:else}
					<p class="text-center text-sm text-muted-foreground">まだクリックデータがありません</p>
				{/if}
			</div>
		{/if}
		<Dialog.Footer>
			<Button variant="secondary" onclick={() => (showStats = false)}>閉じる</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- AI Summary Dialog -->
<Dialog.Root bind:open={showSummary}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>
				<span class="text-primary">AI</span> URL要約
			</Dialog.Title>
		</Dialog.Header>
		<div class="whitespace-pre-wrap text-sm leading-relaxed text-muted-foreground">
			{summaryText}
		</div>
		<Dialog.Footer>
			<Button variant="secondary" onclick={() => (showSummary = false)}>閉じる</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
