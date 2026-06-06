package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

func main() {
	email := flag.String("email", "", "Twickets account email")
	password := flag.String("password", "", "Twickets account password")
	listingURL := flag.String("url", "", "Twickets listing URL to add to cart")
	flag.Parse()

	if *email == "" || *password == "" || *listingURL == "" {
		log.Fatal("Usage: carttest -email=you@example.com -password=secret -url=https://www.twickets.live/app/block/...")
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	// Step 1: log in
	log.Println("Navigating to login page...")
	if _, err = page.Goto("https://www.twickets.live/app/login", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("could not navigate to login: %v", err)
	}

	if err = page.Fill("input[type=email], input[name=email], input[id=email]", *email); err != nil {
		log.Fatalf("could not fill email: %v", err)
	}
	if err = page.Fill("input[type=password], input[name=password], input[id=password]", *password); err != nil {
		log.Fatalf("could not fill password: %v", err)
	}
	if err = page.Click("button[type=submit], input[type=submit], button:has-text('Log in'), button:has-text('Sign in')"); err != nil {
		log.Fatalf("could not click login: %v", err)
	}

	log.Println("Waiting for login to complete...")
	time.Sleep(3 * time.Second)
	log.Printf("Current URL after login: %s\n", page.URL())

	// Step 2: navigate to the listing
	log.Printf("Navigating to listing: %s\n", *listingURL)
	if _, err = page.Goto(*listingURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("could not navigate to listing: %v", err)
	}

	log.Printf("Listing page loaded: %s\n", page.URL())
	time.Sleep(2 * time.Second)

	// Step 3: click buy/add to cart
	log.Println("Looking for buy/add-to-cart button...")
	buySelectors := []string{
		"button:has-text('Buy')",
		"button:has-text('Add to cart')",
		"button:has-text('Purchase')",
		"a:has-text('Buy')",
		"[data-testid='buy-button']",
	}

	clicked := false
	for _, sel := range buySelectors {
		count, err := page.Locator(sel).Count()
		if err == nil && count > 0 {
			log.Printf("Found button with selector: %s\n", sel)
			if err = page.Click(sel); err != nil {
				log.Printf("Could not click %s: %v\n", sel, err)
				continue
			}
			clicked = true
			break
		}
	}

	if !clicked {
		log.Println("Could not find buy button — check the browser window")
	}

	// Wait and observe what happens (CAPTCHA, success, etc.)
	log.Println("Waiting 30 seconds to observe result (check browser window)...")
	time.Sleep(30 * time.Second)

	finalURL := page.URL()
	fmt.Printf("\nFinal URL: %s\n", finalURL)

	if finalURL != *listingURL {
		fmt.Println("Page navigated — likely triggered a CAPTCHA or checkout flow.")
	}
}
