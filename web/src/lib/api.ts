export interface URL {
	code: string;
	original_url: string;
	created_at: string;
	clicks: number;
}

export interface ShortenResponse {
	code: string;
	short_url: string;
}

export async function shortenUrl(url: string): Promise<ShortenResponse> {
	const res = await fetch('/api/shorten', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ url })
	});
	if (!res.ok) {
		const err = await res.json();
		throw new Error(err.error || 'Failed to shorten URL');
	}
	return res.json();
}

export async function listUrls(): Promise<URL[]> {
	const res = await fetch('/api/urls');
	if (!res.ok) throw new Error('Failed to fetch URLs');
	return res.json();
}

export async function deleteUrl(code: string): Promise<void> {
	const res = await fetch(`/api/urls/${code}`, { method: 'DELETE' });
	if (!res.ok) throw new Error('Failed to delete URL');
}

export async function summarizeUrl(code: string): Promise<string> {
	const res = await fetch(`/api/urls/${code}/summarize`, { method: 'POST' });
	if (!res.ok) throw new Error('Failed to summarize URL');
	const data = await res.json();
	return data.summary;
}

export interface DailyClicks {
	date: string;
	clicks: number;
}

export interface ClickStats {
	code: string;
	total_clicks: number;
	daily: DailyClicks[];
}

export async function getClickStats(code: string, days = 30): Promise<ClickStats> {
	const res = await fetch(`/api/urls/${code}/stats?days=${days}`);
	if (!res.ok) throw new Error('Failed to fetch stats');
	return res.json();
}
