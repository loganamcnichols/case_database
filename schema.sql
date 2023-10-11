CREATE TABLE cases(
  id SERIAL PRIMARY key,
  pacer_id INT,
  court_id VARCHAR(4),
  title TEXT,
  case_number VARCHAR(20)
  );

INSERT INTO cases (pacer_id, court_id, title, case_number) VALUES (1320666, 'azd', '2:22-cv-02189-SRB Stanley v. Quintairos Prieto Wood & Boyer PA', '2:22-cv-2189');

CREATE TABLE users (id SERIAL PRIMARY KEY, email, password);