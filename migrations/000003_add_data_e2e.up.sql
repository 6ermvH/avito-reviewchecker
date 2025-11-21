INSERT INTO teams (name) VALUES
  ('team-alpha'),
  ('team-beta'),
  ('team-gamma')
ON CONFLICT (name) DO NOTHING;

INSERT INTO users (id, team_name, username, is_active)
VALUES
  ('alpha-1', 'team-alpha', 'Alpha A', true),
  ('alpha-2', 'team-alpha', 'Alpha B', true),
  ('alpha-3', 'team-alpha', 'Alpha C', false),
  ('beta-1', 'team-beta', 'Beta A', true),
  ('beta-2', 'team-beta', 'Beta B', true),
  ('beta-3', 'team-beta', 'Beta C', true),
  ('gamma-1', 'team-gamma', 'Gamma A', true),
  ('gamma-2', 'team-gamma', 'Gamma B', true),
  ('gamma-3', 'team-gamma', 'Gamma C', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO pull_requests (id, name, author_id, status)
VALUES
  ('e2e-pr-1', 'Prepare fixtures', 'alpha-1', 'OPEN'),
  ('e2e-pr-2', 'Add metrics', 'beta-1', 'MERGED')
ON CONFLICT (id) DO NOTHING;

INSERT INTO pull_request_reviewers (pull_request_id, slot, reviewer_id)
VALUES
  ('e2e-pr-1', 1, 'alpha-2'),
  ('e2e-pr-1', 2, 'alpha-3'),
  ('e2e-pr-2', 1, 'beta-2')
ON CONFLICT (pull_request_id, slot) DO NOTHING;
