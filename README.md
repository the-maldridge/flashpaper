# Flashpaper

Flashpaper is an ephemeral paste service useful for when you need to
transfer some kind of secure data such as a password, ssh key, or
authorization token to someone else.

It's inspired by similar projects that provide quick paste services to
move around secure data.

Flashpaper needs a Redis compatible server running which can be
pointed to by setting `REDIS_ADDR` to the address of your Redis
server.
