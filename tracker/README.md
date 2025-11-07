# Kaunta Tracking Script

Privacy-first analytics tracker combining Umami and Plausible features.

## Features

### Core Tracking
- Auto-track pageviews on initial load
- SPA navigation tracking (pushState/replaceState/popState)
- Custom event tracking with properties
- Session-aware referrer handling
- Screen resolution tracking
- Language detection
- Domain whitelisting

### Advanced Features
- Scroll depth tracking
- Engagement time tracking
- Outbound link tracking (automatic)
- Visibility change handling
- bfcache support
- Dynamic content height recalculation

### Privacy & Performance
- Respects Do Not Track
- No cookies
- No localStorage
- Silent fail on errors
- Minimal size (<3KB minified)
- Uses `fetch` with `keepalive`

## Installation

Add to your HTML `<head>`:

```html
<script
  defer
  data-website-id="your-uuid-here"
  data-api-url="https://census.yourdomain.com"
  src="https://census.yourdomain.com/k.js">
</script>
```

## Configuration

| Attribute | Default | Description |
|-----------|---------|-------------|
| `data-website-id` | required | Your website UUID |
| `data-api-url` | script's directory | API endpoint base URL |
| `data-auto-track` | true | Auto-track pageviews |
| `data-track-outbound` | true | Auto-track outbound link clicks |
| `data-respect-dnt` | true | Respect Do Not Track browser setting |
| `data-exclude-hash` | false | Remove URL hash from tracked URLs |
| `data-domains` | all | Comma-separated list of domains to track |

## Examples

**Minimal setup:**
```html
<script defer data-website-id="550e8400-e29b-41d4-a716-446655440000" src="/k.js"></script>
```

**Disable outbound link tracking:**
```html
<script
  defer
  data-website-id="550e8400-e29b-41d4-a716-446655440000"
  data-track-outbound="false"
  src="/k.js">
</script>
```

**Multi-domain tracking:**
```html
<script
  defer
  data-website-id="550e8400-e29b-41d4-a716-446655440000"
  data-domains="example.com, app.example.com, blog.example.com"
  src="/k.js">
</script>
```

## API Usage

### Custom Events

```javascript
kaunta.track('button_click');

kaunta.track('purchase', {
  product: 'Premium Plan',
  amount: 29.99,
  currency: 'USD'
});

kaunta.track('form_submit', {
  form_name: 'contact',
  fields: 5,
  validation_passed: true
});
```

### Manual Pageviews

```javascript
kaunta.trackPageview();
```

### Examples

**E-commerce:**
```javascript
kaunta.track('add_to_cart', {
  product_id: '12345',
  product_name: 'Blue Widget',
  price: 19.99,
  quantity: 2
});

kaunta.track('purchase', {
  order_id: 'ORD-2024-001',
  total: 45.97,
  currency: 'USD'
});
```

**User engagement:**
```javascript
kaunta.track('video_play', {
  video_id: 'intro-2024',
  duration: 180
});

kaunta.track('newsletter_signup', {
  location: 'footer',
  list: 'weekly-digest'
});

kaunta.track('search', {
  query: 'analytics tools',
  results_count: 24
});
```

**Feature usage:**
```javascript
kaunta.track('cta_click', {
  button_text: 'Start Free Trial',
  location: 'hero'
});

kaunta.track('form_submit', {
  form_id: 'contact-form',
  fields_filled: 5,
  success: true
});
```

## Data Format

Tracker sends data to `/api/send` in Umami-compatible format:

```json
{
  "type": "event",
  "payload": {
    "website": "550e8400-e29b-41d4-a716-446655440000",
    "hostname": "example.com",
    "url": "/products/analytics",
    "title": "Analytics Tools - Example",
    "referrer": "https://google.com",
    "screen": "1920x1080",
    "language": "en-US",
    "scroll_depth": 75,
    "engagement_time": 45000,
    "name": "button_click",
    "props": {
      "button": "signup",
      "location": "header"
    }
  }
}
```

## How It Works

### Pageview Tracking

1. Initial Load: Tracks pageview when script initializes
2. SPA Navigation: Hooks into pushState/replaceState/popstate
3. Hash Changes: Optional hash-based routing (disabled by default)
4. Back/Forward: Captures browser history navigation
5. bfcache: Handles page restoration from browser cache

### Engagement Tracking

**Scroll Depth:**
- Continuously monitors scroll position
- Tracks maximum depth reached
- Calculates percentage based on document height
- Updates on dynamic content changes

**Engagement Time:**
- Starts timer when page becomes visible
- Pauses when tab/window loses focus
- Resumes when focus returns
- Accumulates total engaged time
- Resets on navigation

### Outbound Link Tracking

1. Intercepts all clicks on `<a>` tags
2. Checks if link hostname differs from current hostname
3. Sends analytics event with target URL
4. Delays navigation by 500ms (if safe to intercept)
5. Respects middle-click/cmd-click

## Browser Support

- Chrome/Edge 42+
- Firefox 52+
- Safari 11.1+
- Opera 29+
- iOS Safari 11.3+
- Android Chrome 42+

Graceful degradation for older browsers.

## Privacy Features

### Do Not Track Support

Automatically disables tracking when browser sends DNT header:
- `navigator.doNotTrack === '1'`
- `navigator.doNotTrack === 'yes'`

### No Client-Side Storage

- No cookies
- No localStorage
- No sessionStorage
- No IndexedDB
- No fingerprinting

### GDPR Compliance

Designed for cookieless tracking:
- No consent banner required
- Respects user privacy preferences
- No cross-site tracking
- No personal data collection

## Performance

### Size Budget

- Unminified: ~8KB
- Minified: ~2.8KB
- Minified + Gzip: ~1.2KB

### Optimization Tips

1. Use `defer` attribute (non-blocking script load)
2. CDN/Edge hosting (serve from fast edge locations)
3. Cache headers (set long cache duration)
4. Subresource Integrity (SRI hash for security)

Example with SRI:
```html
<script
  defer
  data-website-id="your-uuid"
  src="/k.js"
  integrity="sha384-YOUR-HASH-HERE"
  crossorigin="anonymous">
</script>
```

### Network Impact

- Single request per pageview (no polling)
- Keepalive flag (survives page unload)
- Silent failures (no retry storms)
- CORS-friendly (works cross-origin)

## Building

### Using Bun (recommended)

```bash
bun install
bun run build:tracker
```

This creates `cmd/kaunta/assets/kaunta.min.js`, which is the version embedded in the Go binary.

### Using Terser

```bash
npm install -g terser

terser kaunta.js \
  --compress \
  --mangle \
  --output ../cmd/kaunta/assets/kaunta.min.js
```

### Using esbuild

```bash
npm install -g esbuild

esbuild kaunta.js \
  --minify \
  --outfile=../cmd/kaunta/assets/kaunta.min.js
```

### Build Script

Run `tracker/build.sh` to invoke Bun and print SRI/size info for the generated bundle.

## Testing

### Browser Console

```javascript
console.log(window.kaunta);
kaunta.track('test_event', { source: 'console' });
kaunta.trackPageview();
```

### Network Monitoring

1. Open DevTools (F12)
2. Network tab
3. Filter by "send"
4. Trigger events
5. Inspect request payloads

## Troubleshooting

### Script Not Tracking

Check if loaded:
```javascript
console.log(window.kaunta);
```

Check configuration:
- Ensure `data-website-id` is set and valid UUID
- Verify `data-api-url` is correct
- Check browser console for errors

Check DNT:
```javascript
console.log(navigator.doNotTrack);
```

### Events Not Appearing

Check domain whitelist:
```javascript
console.log(location.hostname);
```

Check network tab:
- Requests should appear as `POST /api/send`
- Status should be 200 or 202
- Check CORS headers on server if error

Check ad blockers:
- Try in incognito/private mode
- Whitelist your domain

### Scroll/Engagement Not Recording

Check if tracking initialized:
```javascript
kaunta.track('test');
```

Check scroll_depth and engagement_time in Network tab.

## Advanced Usage

### Custom Domain Tracking

Track multiple domains with one website ID:

```javascript
kaunta.track('cross_domain_event', {
  source_domain: 'domain1.com'
});
```

### Tracking User Journeys

```javascript
kaunta.track('user_journey', {
  step: 'landing',
  source: 'google'
});

kaunta.track('user_journey', {
  step: 'signup_started'
});

kaunta.track('user_journey', {
  step: 'signup_completed',
  plan: 'free'
});
```

### A/B Testing

```javascript
var variant = Math.random() < 0.5 ? 'A' : 'B';

kaunta.track('page_view', { variant: variant });
kaunta.track('button_click', { variant: variant, button: 'cta' });
kaunta.track('conversion', { variant: variant, value: 29.99 });
```

## License

MIT License - See Kaunta repository for full license.
