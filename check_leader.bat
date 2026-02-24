@echo off
powershell -NoProfile -Command "$ErrorActionPreference='Stop'; $r=Invoke-RestMethod -Uri 'http://localhost:2379/v3/kv/range' -Method Post -Body '{\"key\":\"L3NlcnZpY2UvcG9zdGdyZXMtaGEvbGVhZGVy\"}' -ContentType 'application/json'; [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($r.kvs[0].value))" || exit /b 1
exit /b 0