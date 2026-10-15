INSERT INTO achievements(id, name, description) VALUES
  ('started_first_case', 'Took the Badge', 'Opened your first case file. The city is yours to lose.'),
  ('found_first_evidence', 'Eagle Eye', 'Bagged your first piece of evidence. Nothing on this train hides forever.'),
  ('broke_alibi', 'Alibi Breaker', 'Confronted a suspect with hard proof and watched their story fall apart.'),
  ('solved_last_train', 'Closed the File', 'Named Edwin Marsh and proved it aboard the Meridian Limited.'),
  ('perfect_deduction', 'Sherlock of Graywater', 'Solved the Last Train with every exhibit found and every witness broken. Flawless.')
ON CONFLICT (id) DO NOTHING;
