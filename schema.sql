CREATE TABLE cases(
  id SERIAL PRIMARY key,
  pacer_id INT,
  court_id VARCHAR(4),
  title TEXT,
  case_number VARCHAR(20)
  );

INSERT INTO cases (pacer_id, court_id, title, case_number) VALUES (1320666, 'azd', '2:22-cv-02189-SRB Stanley v. Quintairos Prieto Wood & Boyer PA', '2:22-cv-2189');

CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT, password CHAR(60), credits INT);
INSERT INTO users(email, password, credits) VALUES ('loganamcnichols@gmail.com', 'password', 0);
CREATE TABLE documents (id SERIAL PRIMARY KEY, description TEXT, file TEXT, doc_number INT, case_id INT, pages INT);
CREATE TABLE users_by_documents (user_id INT, doc_id INT);
CREATE INDEX user_id_idx ON users_by_documents (user_id);
CREATE INDEX doc_id_idx ON users_by_documents (doc_id);

INSERT INTO documents (description, file, doc_number, case_id, pages) VALUES ('CORPORATE DISCLOSURE STATEMENT', '1320666-2.pdf', 2, 1, 3);
INSERT INTO users_by_documents (user_id, doc_id) VALUES (1, 1);