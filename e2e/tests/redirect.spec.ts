import { test, expect, request as playwrightRequest } from "@playwright/test";

const API = "http://localhost:8080";
const DYNAMODB = "http://localhost:8000";

// Helper: create a short URL via API
async function createShortURL(
  request: any,
  originalURL: string
): Promise<{ code: string; short_url: string }> {
  const res = await request.post(`${API}/api/shorten`, {
    data: { url: originalURL },
  });
  expect(res.ok()).toBeTruthy();
  return res.json();
}

// Helper: mark a URL as unsafe via DynamoDB Local
async function markUnsafe(code: string) {
  const ctx = await playwrightRequest.newContext();
  const res = await ctx.post(DYNAMODB, {
    headers: {
      "Content-Type": "application/x-amz-json-1.0",
      "X-Amz-Target": "DynamoDB_20120810.UpdateItem",
      Authorization:
        "AWS4-HMAC-SHA256 Credential=dummy/20260101/ap-northeast-1/dynamodb/aws4_request, SignedHeaders=content-type;host;x-amz-target, Signature=dummy",
    },
    data: {
      TableName: "url-shortener",
      Key: { code: { S: code } },
      UpdateExpression: "SET safe_status = :s",
      ExpressionAttributeValues: { ":s": { S: "unsafe" } },
    },
  });
  await ctx.dispose();
  return res;
}

test.describe("Redirect - safe URL", () => {
  test("safe URL redirects with 301", async ({ request }) => {
    const { code } = await createShortURL(request, "https://example.com");

    const res = await request.get(`${API}/r/${code}`, {
      maxRedirects: 0,
    });
    expect(res.status()).toBe(301);
    expect(res.headers()["location"]).toBe("https://example.com");
  });

  test("non-existent code returns 404", async ({ request }) => {
    const res = await request.get(`${API}/r/nonexistent`, {
      maxRedirects: 0,
    });
    expect(res.status()).toBe(404);
  });
});

test.describe("Redirect - unsafe URL warning page", () => {
  test("unsafe URL shows warning page instead of redirect", async ({
    page,
    request,
  }) => {
    const { code } = await createShortURL(request, "https://example.com");

    // Mark as unsafe via DynamoDB Local
    const updateRes = await markUnsafe(code);
    expect(updateRes.ok()).toBeTruthy();

    // Visit the redirect URL — should show warning page
    await page.goto(`${API}/r/${code}`);

    await expect(page.locator("h1")).toContainText("Warning");
    await expect(page.locator("body")).toContainText("potentially unsafe");
    await expect(page.locator("code")).toContainText("https://example.com");

    // "Continue anyway" link should exist and point to original URL
    const continueLink = page.locator('a:has-text("Continue anyway")');
    await expect(continueLink).toBeVisible();
    await expect(continueLink).toHaveAttribute(
      "href",
      "https://example.com"
    );
  });
});

test.describe("Health endpoint", () => {
  test("returns ok status", async ({ request }) => {
    const res = await request.get(`${API}/health`);
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.status).toBe("ok");
  });
});
