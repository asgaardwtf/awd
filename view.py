import undetected_chromedriver.v2 as uc
from fake_useragent import UserAgent
import threading
import argparse
import random
import time
from selenium.webdriver.common.by import By

# Function to set custom headers
def add_custom_headers(driver, custom_headers):
    for key, value in custom_headers.items():
        driver.execute_cdp_cmd('Network.setExtraHTTPHeaders', {'headers': {key: value}})

# Function to view a webpage with random user-agent and additional headers
def view_page(url, headers):
    options = uc.ChromeOptions()

    # Generate a random user-agent
    ua = UserAgent()
    user_agent = ua.random
    options.add_argument(f'user-agent={user_agent}')
    
    # Disable WebRTC to avoid detection
    options.add_argument('--disable-webrtc')

    # Start undetected Chrome with custom options
    driver = uc.Chrome(options=options)

    # Add custom headers
    add_custom_headers(driver, headers)

    try:
        # Access the page
        driver.get(url)
        print(f"Successfully accessed {url} with User-Agent: {user_agent}")

        # Simulate browsing delay
        time.sleep(random.randint(5, 10))

        # Example of finding an element (you can change this based on your needs)
        # driver.find_element(By.TAG_NAME, 'body')

    except Exception as e:
        print(f"Error accessing {url}: {e}")

    finally:
        driver.quit()

# Threading function to handle concurrency
def run_threads(urls, num_threads, headers):
    threads = []
    for url in urls:
        # Create a thread for each URL
        t = threading.Thread(target=view_page, args=(url, headers))
        threads.append(t)

        if len(threads) >= num_threads:
            # Start the threads
            for thread in threads:
                thread.start()
            # Wait for all threads to finish
            for thread in threads:
                thread.join()
            # Clear thread list for the next batch
            threads = []

# Main function for argument parsing and running the program
def main():
    parser = argparse.ArgumentParser(description="Web Viewer using undetected-chromedriver")
    parser.add_argument('-u', '--urls', nargs='+', required=True, help='List of URLs to visit')
    parser.add_argument('-t', '--threads', type=int, default=5, help='Number of concurrent threads')
    parser.add_argument('-c', '--custom-headers', nargs='*', help='Custom headers in key:value format', default=[])

    args = parser.parse_args()

    # Prepare custom headers dictionary
    headers = {}
    for header in args.custom_headers:
        key, value = header.split(":")
        headers[key.strip()] = value.strip()

    # Run the threads to visit each URL concurrently
    run_threads(args.urls, args.threads, headers)

if __name__ == "__main__":
    main()
