import argparse
from undetected_chromedriver.v2 import Chrome, ChromeOptions
import threading
from fake_useragent import UserAgent
import random

def view_url(url):
    options = ChromeOptions()
    ua = UserAgent()
    options.add_argument(f"user-agent={ua.random}")
    options.add_argument("start-maximized")
    options.add_argument("disable-infobars")
    options.add_argument("--disable-extensions")
    options.add_argument("--disable-dev-shm-usage")
    options.add_argument("--disable-browser-side-navigation")
    options.add_argument("--disable-gpu")
    options.add_argument("--no-sandbox")
    options.add_argument("--ignore-certificate-errors")
    options.add_argument("--disable-notifications")
    options.add_argument("--disable-features=VizDisplayCompositor")
    options.add_argument(f"--window-position={random.randint(0,1920)},{random.randint(0,1080)}")
    options.add_argument(f"--user-agent={ua.random}")

    driver = Chrome(options=options)
    driver.header_overrides = {"X-Geo": "DK"}
    driver.get(url)
    driver.quit()

def main():
    parser = argparse.ArgumentParser(description='Automatically view URLs with specific headers and concurrency')
    parser.add_argument('--concurrency', type=int, default=2, help='Number of concurrent threads')
    parser.add_argument('--url', type=str, required=True, help='URL to view')
    args = parser.parse_args()

    urls = [args.url] * args.concurrency

    threads = []
    for url in urls:
        t = threading.Thread(target=view_url, args=(url,))
        threads.append(t)
        t.start()

    for t in threads:
        t.join()

if __name__ == '__main__':
    main()
