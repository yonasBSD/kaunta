-- Migration 000008 Rollback: Remove Pagination Support

-- Drop the new get_top_pages function
DROP FUNCTION IF EXISTS get_top_pages(UUID, INTEGER, INTEGER, INTEGER, VARCHAR, VARCHAR, VARCHAR);

-- Restore old get_breakdown function (without offset and total_count)
DROP FUNCTION IF EXISTS get_breakdown(UUID, VARCHAR, INTEGER, INTEGER, INTEGER, VARCHAR, VARCHAR, VARCHAR, VARCHAR);

CREATE OR REPLACE FUNCTION get_breakdown(
    p_website_id UUID,
    p_dimension VARCHAR,
    p_days INTEGER DEFAULT 1,
    p_limit INTEGER DEFAULT 10,
    p_country VARCHAR DEFAULT NULL,
    p_browser VARCHAR DEFAULT NULL,
    p_device VARCHAR DEFAULT NULL,
    p_page_path VARCHAR DEFAULT NULL
)
RETURNS TABLE (name VARCHAR, count BIGINT) AS $$
BEGIN
    CASE p_dimension
        WHEN 'country' THEN
            RETURN QUERY
            SELECT COALESCE(s.country, 'Unknown')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND (p_browser IS NULL OR s.browser = p_browser)
              AND (p_device IS NULL OR s.device = p_device)
              AND (p_page_path IS NULL OR e.url_path = p_page_path)
            GROUP BY s.country
            ORDER BY count DESC
            LIMIT p_limit;

        WHEN 'browser' THEN
            RETURN QUERY
            SELECT COALESCE(s.browser, 'Unknown')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND (p_country IS NULL OR s.country = p_country)
              AND (p_device IS NULL OR s.device = p_device)
              AND (p_page_path IS NULL OR e.url_path = p_page_path)
            GROUP BY s.browser
            ORDER BY count DESC
            LIMIT p_limit;

        WHEN 'device' THEN
            RETURN QUERY
            SELECT COALESCE(s.device, 'Unknown')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND (p_country IS NULL OR s.country = p_country)
              AND (p_browser IS NULL OR s.browser = p_browser)
              AND (p_page_path IS NULL OR e.url_path = p_page_path)
            GROUP BY s.device
            ORDER BY count DESC
            LIMIT p_limit;

        WHEN 'referrer' THEN
            RETURN QUERY
            SELECT COALESCE(e.referrer_domain, 'Direct / None')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND (p_country IS NULL OR s.country = p_country)
              AND (p_browser IS NULL OR s.browser = p_browser)
              AND (p_device IS NULL OR s.device = p_device)
              AND (p_page_path IS NULL OR e.url_path = p_page_path)
            GROUP BY e.referrer_domain
            ORDER BY count DESC
            LIMIT p_limit;

        WHEN 'city' THEN
            RETURN QUERY
            SELECT COALESCE(s.city, 'Unknown')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND (p_country IS NULL OR s.country = p_country)
              AND (p_browser IS NULL OR s.browser = p_browser)
              AND (p_device IS NULL OR s.device = p_device)
              AND (p_page_path IS NULL OR e.url_path = p_page_path)
            GROUP BY s.city
            ORDER BY count DESC
            LIMIT p_limit;

        WHEN 'region' THEN
            RETURN QUERY
            SELECT COALESCE(s.region, 'Unknown')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND (p_country IS NULL OR s.country = p_country)
              AND (p_browser IS NULL OR s.browser = p_browser)
              AND (p_device IS NULL OR s.device = p_device)
              AND (p_page_path IS NULL OR e.url_path = p_page_path)
            GROUP BY s.region
            ORDER BY count DESC
            LIMIT p_limit;

        WHEN 'page' THEN
            RETURN QUERY
            SELECT COALESCE(e.url_path, 'Unknown')::VARCHAR as name, COUNT(*)::BIGINT as count
            FROM website_event e
            JOIN session s ON e.session_id = s.session_id
            WHERE e.website_id = p_website_id
              AND e.created_at >= CURRENT_DATE - (p_days || ' days')::INTERVAL
              AND e.event_type = 1
              AND e.url_path IS NOT NULL
              AND (p_country IS NULL OR s.country = p_country)
              AND (p_browser IS NULL OR s.browser = p_browser)
              AND (p_device IS NULL OR s.device = p_device)
            GROUP BY e.url_path
            ORDER BY count DESC
            LIMIT p_limit;

        ELSE
            RAISE EXCEPTION 'Invalid dimension: %. Must be country, browser, device, referrer, city, region, or page', p_dimension;
    END CASE;
END;
$$ LANGUAGE plpgsql STABLE;
