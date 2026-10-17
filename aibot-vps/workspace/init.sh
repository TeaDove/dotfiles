#!/usr/bin/env bash

ls -la /run/secrets/*
scp /run/secrets/* worker-server:/tmp
