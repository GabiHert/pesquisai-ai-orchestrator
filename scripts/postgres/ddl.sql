create schema pesquisai;

create table pesquisai.requests (

                                   id UUID primary key,

                                   context VARCHAR(1000),

                                   research VARCHAR(1000),

                                   total_researches INT default 0,

                                   total_finished_researches INT default 0,

                                   status VARCHAR(10) DEFAULT 'PENDING',

                                   overall TEXT,

                                   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                   updated_at TIMESTAMP

);



create table pesquisai.researches (

                                    id UUID primary key,

                                    request_id UUID ,

                                    title VARCHAR(50),

                                    link VARCHAR(100),

                                    status VARCHAR(10) DEFAULT 'PROCESSING',

                                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

                                    updated_at TIMESTAMP,

                                    summary TEXT,

                                    foreign key (request_id) references pesquisai.requests(id)


);


create table pesquisai.location (

                                    id UUID primary key,

                                    name VARCHAR(10)

);


create table pesquisai.requests_location (

                                            requests_id UUID,

                                            location_id UUID,

                                            primary key (requests_id,location_id),

                                            foreign key (requests_id) references pesquisai.requests(id),

                                            foreign key (location_id) references pesquisai.location(id)

);



create table pesquisai.language (

                                    id UUID primary key,

                                    name VARCHAR(10)

);



create table pesquisai.requests_language (

                                            requests_id UUID,

                                            language_id UUID,

                                            primary key (requests_id,language_id),

                                            foreign key (requests_id) references pesquisai.requests(id),

                                            foreign key (language_id) references pesquisai.language(id)

);

