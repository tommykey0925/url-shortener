import { test, expect } from "@playwright/test";

test.describe("URL shortening flow", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.evaluate(() => localStorage.clear());
    await page.reload();
    // Wait for SPA to hydrate (ssr=false)
    await page.waitForLoadState("networkidle");
    await expect(page.locator('input[type="url"]')).toBeVisible({
      timeout: 10_000,
    });
  });

  test("page loads with form and empty state", async ({ page }) => {
    await expect(page.locator('input[type="url"]')).toBeVisible();
    await expect(page.locator('button:has-text("短縮する")')).toBeVisible();
    await expect(page.locator("text=まだURLがありません")).toBeVisible();
  });

  test("shorten a URL and see it in the list", async ({ page }) => {
    await page.fill('input[type="url"]', "https://example.com");
    await page.click('button:has-text("短縮する")');

    // Wait for result card
    await expect(page.locator("text=短縮URLが作成されました")).toBeVisible({
      timeout: 10_000,
    });

    // Short URL link should appear
    const shortUrlLink = page.locator('a[href*="/r/"]').first();
    await expect(shortUrlLink).toBeVisible();

    // Copy button should exist
    await expect(page.locator('button:has-text("コピー")')).toBeVisible();

    // URL should appear in the list
    await expect(page.locator("text=example.com").first()).toBeVisible();
  });

  test("shows error for invalid URL", async ({ page }) => {
    await page.fill(
      'input[type="url"]',
      "https://this-domain-does-not-exist-xyzzy.example"
    );
    await page.click('button:has-text("短縮する")');

    // Should show error
    const errorText = page.locator("text=unsafe URL");
    await expect(errorText).toBeVisible({ timeout: 15_000 });
  });

  test("delete a URL from the list", async ({ page }) => {
    // Create a URL first
    await page.fill('input[type="url"]', "https://example.com");
    await page.click('button:has-text("短縮する")');
    await expect(page.locator("text=短縮URLが作成されました")).toBeVisible({
      timeout: 10_000,
    });

    // URL should be in the list
    await expect(page.locator("text=example.com").first()).toBeVisible();

    // Click delete button
    await page.click('button:has-text("削除")');

    // URL should disappear from list, empty state should show
    await expect(page.locator("text=まだURLがありません")).toBeVisible({
      timeout: 5_000,
    });
  });
});
