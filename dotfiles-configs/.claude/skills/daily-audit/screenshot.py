import sys

from playwright.sync_api import sync_playwright

DASHBOARD_URL = "http://127.0.0.1:3000/d/homeobs/home-observability?kiosk&from=now-24h&to=now"
VIEWPORT_WIDTH = 1600
SCROLL_STEP = 800
TARGET_CHUNK_HEIGHT = 2000
ROW_SELECTOR = '[data-testid*="dashboard-row"]'
ROW_CONTAINER_MIN_HEIGHT = 50


def row_boundaries(page) -> list[float]:
    rows = page.locator(ROW_SELECTOR)
    tops = []

    for i in range(rows.count()):
        box = rows.nth(i).bounding_box()
        if box and box["height"] > ROW_CONTAINER_MIN_HEIGHT:
            tops.append(box["y"])

    return sorted(set(tops))


def chunk_boundaries(tops: list[float], full_height: float) -> list[float]:
    if not tops:
        return [0, full_height]

    boundaries = [0.0]
    chunk_start = tops[0]

    for top in tops[1:]:
        if top - chunk_start >= TARGET_CHUNK_HEIGHT:
            boundaries.append(top)
            chunk_start = top

    boundaries.append(full_height)
    return boundaries


def main() -> None:
    output_prefix = sys.argv[1]
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

        boundaries = chunk_boundaries(row_boundaries(page), full_height)

        paths = []
        for i in range(len(boundaries) - 1):
            top, bottom = boundaries[i], boundaries[i + 1]
            path = f"{output_prefix}-{i + 1}.png"
            page.screenshot(
                path=path,
                clip={"x": 0, "y": top, "width": VIEWPORT_WIDTH, "height": bottom - top},
            )
            paths.append(path)

        browser.close()

    for path in paths:
        print(path)


if __name__ == "__main__":
    main()
