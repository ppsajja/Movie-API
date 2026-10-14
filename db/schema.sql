-- Run this in Supabase SQL Editor
create table if not exists movies (
  id          bigint generated always as identity primary key,
  title       text not null,
  year        int,
  genre       text,
  director    text,
  rating      numeric(3,1),
  poster_url  text,
  description text,
  created_at  timestamptz not null default now()
);

insert into movies (title, year, genre, director, rating, poster_url, description) values
('The Shawshank Redemption', 1994, 'Drama', 'Frank Darabont', 9.3, 'https://image.tmdb.org/t/p/w500/9cqNxx0GxF0bflZmeSMuL5tnGzr.jpg', 'Two imprisoned men bond over a number of years.'),
('The Godfather', 1972, 'Crime', 'Francis Ford Coppola', 9.2, 'https://image.tmdb.org/t/p/w500/3bhkrj58Vtu7enYsRolD1fZdja1.jpg', 'The aging patriarch of an organized crime dynasty transfers control to his son.'),
('The Dark Knight', 2008, 'Action', 'Christopher Nolan', 9.0, 'https://image.tmdb.org/t/p/w500/qJ2tW6WMUDux911r6m7haRef0WH.jpg', 'Batman faces the Joker.'),
('Inception', 2010, 'Sci-Fi', 'Christopher Nolan', 8.8, 'https://image.tmdb.org/t/p/w500/9gk7adHYeDvHkCSEqAvQNLV5Uge.jpg', 'A thief who steals corporate secrets through dream-sharing technology.'),
('Spirited Away', 2001, 'Animation', 'Hayao Miyazaki', 8.6, 'https://image.tmdb.org/t/p/w500/39wmItIWsg5sZMyRUHLkWBcuVCM.jpg', 'A girl wanders into a world of spirits.'),
('Parasite', 2019, 'Thriller', 'Bong Joon-ho', 8.5, 'https://image.tmdb.org/t/p/w500/7IiTTgloJzvGI1TAYymCfbfl3vT.jpg', 'Greed and class discrimination threaten a symbiotic relationship.');
