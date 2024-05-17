# Deploylock

Simple distributed locks using [lockgate](https://github.com/werf/lockgate).

## Deploylock Server

The server holds the locks.  It can be deployed to Heroku with [git push](https://devcenter.heroku.com/articles/git).

## Deploylock Client

The client is a command-line application that can request a lock and renew a lock.

Use `deploylock-client acquire -n <lockname> https://deploylock-server/` to acquire a lock.
It will block until the lock is successfully retrieved.  Clients waiting to
acquire the lock are placed in a FIFO queue.

Upon successfully acquiring the lock, the lock's UUID will be written to stdout
and `deploylock-client` will exit.

Keep the lock by running `deploylock-client renew -n <lockname -l <uuid>
https://deploylock-server/`.  This can be placed in the background to run other
commands that require the lock.  Kill `deploylock-client renew` to release the
lock.

## Motivation

Deploylock was built to help coordinate long-running builds against Salesforce
orgs.  Due to defects in the Metadata API, it's sometimes necessary to perform
multiple deployments sequentially.  Deploylock can be used to prevent
deployments from independent builds from getting intertwined by using an
external lock on the org.

## Limitations

Shared locks aren't supported.  Lockgate supports shared locks, but a use case
for deploylock hasn't been identified yet.
