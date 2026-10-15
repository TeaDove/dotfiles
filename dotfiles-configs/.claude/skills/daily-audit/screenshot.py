import sys

from playwright.sync_api import sync_playwright

DASHBOARD_URL = "http://127.0.0.1:3000/d/homeobs/home-observability?kiosk&from=now-24h&to=now"
VIEWPORT_WIDTH = 1600
SCROLL_STEP = 800


def main() -> None:
    output_path = sys.argv[1]
    token = sys.stdin.read().strip()

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch()
        context = browser.new_context(
            viewport={"width": VIEWPORT_WIDTH, "height": 1200},
            extra_http_headers={"Authorization": f"Bearer {token}"},
        )
        page = context.new_page()
        page.goto(DASHBOARD_URL, wait_until="networkidle", timeout=30000)
        page.wait_for_timeout(2000)

        full_height = page.evaluate("document.body.scrollHeight")
        page.set_viewport_size({"width": VIEWPORT_WIDTH, "height": full_height})
        page.wait_for_timeout(1000)

        y = 0
        while y < full_height:
            page.evaluate(f"window.scrollTo(0, {y})")
            page.wait_for_timeout(250)
            y += SCROLL_STEP
        page.evaluate("window.scrollTo(0, 0)")
        page.wait_for_timeout(4000)

        page.screenshot(path=output_path, full_page=True)
        browser.close()


if __name__ == "__main__":
    main()
