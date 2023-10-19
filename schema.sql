CREATE TABLE cases(
  id SERIAL PRIMARY key,
  pacer_id INT,
  court_id VARCHAR(4),
  title TEXT,
  case_number VARCHAR(20)
  );
CREATE UNIQUE INDEX pacer_case_idx ON cases (pacer_id, court_id);

INSERT INTO cases (pacer_id, court_id, title, case_number) VALUES (1320666, 'azd', '2:22-cv-02189-SRB Stanley v. Quintairos Prieto Wood & Boyer PA', '2:22-cv-2189');

CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT UNIQUE, password CHAR(60), credits INT);
INSERT INTO users(email, password, credits) VALUES ('loganamcnichols@gmail.com', '$2a$10$9kdF2hCROwNziWZvvMyPiOhDNEi5OWXrpus6exMxj70KwpyAXjauS', 0);
CREATE TABLE documents (id SERIAL PRIMARY KEY, description TEXT, file TEXT, doc_number INT, case_id INT, pages INT, user_id INT, pacer_id TEXT, court TEXT);
CREATE UNIQUE INDEX pacer_doc_idx ON documents (pacer_id, court);
CREATE TABLE users_by_documents (user_id INT, doc_id INT);
CREATE INDEX user_id_idx ON users_by_documents (user_id);
CREATE INDEX doc_id_idx ON users_by_documents (doc_id);

INSERT INTO documents (description, file, doc_number, case_id, pages, user_id, pacer_id, court) VALUES ('CORPORATE DISCLOSURE STATEMENT', '1320666-2.pdf', 2, 1320666, 3, 1, '025125809493', 'azd');
INSERT INTO users_by_documents (user_id, doc_id) VALUES (1, 1);