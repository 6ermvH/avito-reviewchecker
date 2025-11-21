DELETE FROM pull_request_reviewers WHERE pull_request_id IN ('e2e-pr-1', 'e2e-pr-2');
DELETE FROM pull_requests WHERE id IN ('e2e-pr-1', 'e2e-pr-2');
DELETE FROM users WHERE id IN ('alpha-1','alpha-2','alpha-3','beta-1','beta-2','beta-3','gamma-1','gamma-2','gamma-3');
DELETE FROM teams WHERE name IN ('team-alpha','team-beta','team-gamma');
