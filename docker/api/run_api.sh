#!/bin/sh

echo 'Runing migrations...'
migrate up

echo 'Start application...'
api