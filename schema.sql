CREATE TABLE cases(
  id SERIAL PRIMARY key,
  pacer_id INT,
  court_id VARCHAR(4),
  title TEXT,
  case_number VARCHAR(20)
  );

INSERT INTO cases (pacer_id, court_id, title, case_number) VALUES (1320666, 'azd', '2:22-cv-02189-SRB Stanley v. Quintairos Prieto Wood & Boyer PA', '2:22-cv-2189');

CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT, password CHAR(60));

CREATE TABLE documents (id SERIAL PRIMARY KEY, description TEXT, file TEXT, doc_number INT, case_id INT);
CREATE TABLE users_by_documents (user_id INT, doc_id INT);
CREATE INDEX user_id_idx ON users_by_documents (user_id);
CREATE INDEX doc_id_idx ON users_by_documents (doc_id);

INSERT INTO documents (description, file, doc_number, case_id) VALUES ('CORPORATE DISCLOSURE STATEMENT', '1320666-2.pdf', 2, 1);
INSERT INTO users_by_documents (user_id, doc_id) VALUES (5, 1);