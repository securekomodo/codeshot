SELECT
  u.handle,
  count(s.id)        AS snippets,
  max(s.created_at)  AS last_shot
FROM users u
JOIN snippets s ON s.owner_id = u.id
WHERE s.public AND s.created_at > now() - interval '30 days'
GROUP BY u.handle
ORDER BY snippets DESC
LIMIT 10;
