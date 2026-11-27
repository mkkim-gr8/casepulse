-- +goose Up
-- 사건 기본 정보입니다. 법원은 Leviathan이 소유하므로 외래 키 없이 ID만 저장합니다.
CREATE TABLE cases (
    id          uuid        PRIMARY KEY DEFAULT uuidv7(),
    year        integer     NOT NULL CHECK (year BETWEEN 1000 AND 9999),
    code        text        NOT NULL CHECK (code <> ''),
    serial      integer     NOT NULL CHECK (serial >= 1),
    title       text        NOT NULL CHECK (title <> ''),
    category    text        NOT NULL CHECK (category IN ('civil', 'administrative', 'family', 'civil_application', 'criminal')),
    court_id    text        NOT NULL CHECK (court_id <> ''),
    panel_name  text        NOT NULL CHECK (panel_name <> ''),
    created_at  timestamptz NOT NULL DEFAULT now(),
    -- 사건번호는 법원마다 따로 매기므로 법원과 함께 있어야 고유합니다.
    CONSTRAINT cases_court_number_key UNIQUE (court_id, year, code, serial)
);

-- 목록은 최근 등록 순, 같은 때 등록한 사건은 사건번호가 큰 순으로 보여 줍니다.
CREATE INDEX cases_list_order_idx ON cases (created_at DESC, year DESC, serial DESC);

-- 사건 당사자입니다. 입력 순서를 position으로 유지합니다.
CREATE TABLE case_parties (
    case_id   uuid    NOT NULL REFERENCES cases (id) ON DELETE CASCADE,
    position  integer NOT NULL CHECK (position >= 0),
    role      text    NOT NULL CHECK (role <> ''),
    name      text    NOT NULL CHECK (name <> ''),
    client    boolean NOT NULL DEFAULT false,
    PRIMARY KEY (case_id, position)
);

-- +goose Down
DROP TABLE case_parties;
DROP TABLE cases;
