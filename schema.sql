CREATE TABLE cases(
  id SERIAL PRIMARY key,
  pacer_id INT,
  court_id VARCHAR(4),
  title TEXT
  );

INSERT INTO cases (pacer_id, court_id, title) VALUES (1320666, 'azd', '2:22-cv-02189-SRB Stanley v. Quintairos Prieto Wood & Boyer PA');
