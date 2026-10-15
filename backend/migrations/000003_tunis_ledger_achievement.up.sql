INSERT INTO achievements(id, name, description) VALUES
  ('solved_tunis_ledger', 'Harbor Light', 'Named Marco Silvestri and proved it at the Hôtel Majestic.')
ON CONFLICT (id) DO NOTHING;
