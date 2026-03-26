CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),
    username VARCHAR(100) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE quizzes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE questions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    quiz_id BIGINT,
    content TEXT,
    point INT DEFAULT 10,
    time_limit INT DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_quiz_id (quiz_id)
);

CREATE TABLE answers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    question_id BIGINT,
    content TEXT,
    is_correct BOOLEAN,
    INDEX idx_question_id (question_id)
);

CREATE TABLE quiz_sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    quiz_id BIGINT,
    session_code VARCHAR(20) UNIQUE,
    status ENUM('waiting', 'running', 'finished') DEFAULT 'waiting',
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE session_participants (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT,
    user_id BIGINT,
    score INT DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_user_session (session_id, user_id),
    INDEX idx_session (session_id)
);

CREATE TABLE user_answers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT,
    user_id BIGINT,
    question_id BIGINT,
    answer_id BIGINT,
    is_correct BOOLEAN,
    score INT DEFAULT 0,
    answered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uniq_answer (session_id, user_id, question_id),
    INDEX idx_session_user (session_id, user_id),
    INDEX idx_question (question_id)
);

-- Seed Data
INSERT INTO quizzes (title, description) VALUES 
('General Knowledge', 'A fun quiz testing your general knowledge about the world.'),
('Go Programming', 'Test your skills in Go (Golang).');

-- Questions for General Knowledge (Quiz ID 1)
INSERT INTO questions (quiz_id, content, point, time_limit) VALUES 
(1, 'What is the capital of France?', 10, 15),
(1, 'Which planet is known as the Red Planet?', 10, 15),
(1, 'Who wrote "To Kill a Mockingbird"?', 10, 20),
(1, 'What is the largest ocean on Earth?', 10, 15),
(1, 'Which is the tallest mountain in the world?', 10, 15),
(1, 'What is the chemical symbol for Gold?', 10, 10),
(1, 'Which is the largest land animal?', 10, 15),
(1, 'In which year did the Titanic sink?', 10, 15),
(1, 'Who painted the Mona Lisa?', 10, 20),
(1, 'What is the currency of Japan?', 10, 10),
(1, 'Which is the smallest country in the world?', 10, 15),
(1, 'What is the square root of 144?', 10, 10);

-- Answers for General Knowledge (Questions 1-12)
INSERT INTO answers (question_id, content, is_correct) VALUES 
(1, 'Paris', TRUE), (1, 'London', FALSE), (1, 'Berlin', FALSE), (1, 'Madrid', FALSE),
(2, 'Earth', FALSE), (2, 'Mars', TRUE), (2, 'Jupiter', FALSE), (2, 'Venus', FALSE),
(3, 'Harper Lee', TRUE), (3, 'Mark Twain', FALSE), (3, 'Ernest Hemingway', FALSE), (3, 'F. Scott Fitzgerald', FALSE),
(4, 'Atlantic', FALSE), (4, 'Pacific', TRUE), (4, 'Indian', FALSE), (4, 'Arctic', FALSE),
(5, 'K2', FALSE), (5, 'Mount Everest', TRUE), (5, 'Kangchenjunga', FALSE), (5, 'Lhotse', FALSE),
(6, 'Ag', FALSE), (6, 'Au', TRUE), (6, 'Pb', FALSE), (6, 'Fe', FALSE),
(7, 'Elephant', TRUE), (7, 'Giraffe', FALSE), (7, 'Hippo', FALSE), (7, 'Rhino', FALSE),
(8, '1910', FALSE), (8, '1912', TRUE), (8, '1914', FALSE), (8, '1916', FALSE),
(9, 'Picasso', FALSE), (9, 'Da Vinci', TRUE), (9, 'Van Gogh', FALSE), (9, 'Michelangelo', FALSE),
(10, 'Yuan', FALSE), (10, 'Yen', TRUE), (10, 'Won', FALSE), (10, 'Ringgit', FALSE),
(11, 'Monaco', FALSE), (11, 'Vatican City', TRUE), (11, 'San Marino', FALSE), (11, 'Liechtenstein', FALSE),
(12, '10', FALSE), (12, '12', TRUE), (12, '14', FALSE), (12, '16', FALSE);

-- Questions for Go Programming (Quiz ID 2)
INSERT INTO questions (quiz_id, content, point, time_limit) VALUES 
(2, 'Who created Go?', 10, 20),
(2, 'What is the default value of an int in Go?', 10, 10),
(2, 'Which keyword is used to start a goroutine?', 10, 10),
(2, 'How do you iterate over a map in Go?', 10, 15),
(2, 'Does Go have classes?', 10, 10),
(2, 'What is the standard way to handle errors in Go?', 10, 15),
(2, 'Which package is used for formatted I/O?', 10, 10),
(2, 'Which data structure is used for unique elements?', 10, 15),
(2, 'What is the purpose of the init() function?', 10, 20),
(2, 'Is Go a compiled or interpreted language?', 10, 10),
(2, 'How do you define a constant in Go?', 10, 10),
(2, 'What is the default value of a slice?', 10, 10);

-- Answers for Go Programming (Questions 13-24)
INSERT INTO answers (question_id, content, is_correct) VALUES 
(13, 'Google', TRUE), (13, 'Meta', FALSE), (13, 'Microsoft', FALSE), (13, 'Apple', FALSE),
(14, '0', TRUE), (14, '1', FALSE), (14, '-1', FALSE), (14, 'nil', FALSE),
(15, 'go', TRUE), (15, 'start', FALSE), (15, 'run', FALSE), (15, 'thread', FALSE),
(16, 'for', FALSE), (16, 'range', TRUE), (16, 'each', FALSE), (16, 'map.iter', FALSE),
(17, 'Yes', FALSE), (17, 'No', TRUE), (17, 'Only via plugins', FALSE), (17, 'Since v1.18', FALSE),
(18, 'Try-Catch', FALSE), (18, 'Return values', TRUE), (18, 'Exceptions', FALSE), (18, 'Panic only', FALSE),
(19, 'io', FALSE), (19, 'fmt', TRUE), (19, 'log', FALSE), (19, 'print', FALSE),
(20, 'Set', FALSE), (20, 'Map', TRUE), (20, 'Array', FALSE), (20, 'List', FALSE),
(21, 'Cleanup', FALSE), (21, 'Initialization', TRUE), (21, 'Main loop', FALSE), (21, 'Testing', FALSE),
(22, 'Compiled', TRUE), (22, 'Interpreted', FALSE), (22, 'Both', FALSE), (22, 'JIT', FALSE),
(23, 'const', TRUE), (23, 'var', FALSE), (23, 'let', FALSE), (23, 'final', FALSE),
(24, '[]', FALSE), (24, 'nil', TRUE), (24, '0', FALSE), (24, 'empty', FALSE);
