# Safe Categorization Guide

## What Was Fixed

### ✅ Fixed Issue #1: Filter
- **Before**: Categorized ALL playlists (including private ones)
- **After**: Only categorizes PUBLIC playlists without categories

### ✅ Fixed Issue #2: Rate Limiting
- **Before**: Sent all requests at once → Hit 15 requests/minute limit
- **After**: 5-second delay between each request (12 requests/minute max)

## How to Use

### Option 1: Categorize All Public Uncategorized Playlists (Recommended)

```bash
# This will automatically rate-limit to avoid API quota
curl http://localhost:8080/debug/categorize-all
```

**What it does:**
- Finds all PUBLIC playlists with empty categories
- Submits them to the worker queue
- Workers process them with 5-second delays between each
- Safe for free tier (15 requests/minute limit)

### Option 2: Categorize a Single Playlist

```bash
curl http://localhost:8080/debug/categorize/YOUR_PLAYLIST_ID
```

## Expected Behavior

### If you have 10 uncategorized public playlists:
- Time to complete: ~50 seconds (10 × 5 seconds)
- Rate: 12 playlists/minute
- Will NOT hit API limit ✅

### If you have 50+ playlists:
- Time to complete: ~4-5 minutes
- Processes safely without quota errors
- Check console logs to see progress

## Checking Progress

Watch your backend console for logs:
```
Categorizing playlist 550e8400-e29b-41d4-a716-446655440000...
Categorization completed for playlist 550e8400-e29b-41d4-a716-446655440000
```

## If You Still Get Rate Limit Errors

The worker adds a 5-second delay, but if you trigger it multiple times or have other processes, you might still hit limits.

**Solution**: Wait 1 minute, then try again.

## Upgrading API Tier (Optional)

If you need faster categorization:
1. Go to [Google AI Studio](https://aistudio.google.com/)
2. Upgrade to paid tier for higher limits
3. Update your API key if needed
