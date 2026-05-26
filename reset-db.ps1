docker exec ego-postgres-1 psql -U ego -d ego -c "TRUNCATE TABLE chat_messages, chat_sessions, constellations, echos, insights, moments, stars, traces, users CASCADE;"
Write-Host "done - all tables cleared"
