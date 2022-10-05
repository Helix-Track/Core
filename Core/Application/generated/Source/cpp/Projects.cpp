/*
    Projects.cpp
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "Projects.h"

std::string Projects::getId() {    
    return this->id;
}

void Projects::setId(std::string value) {
    this->id = value;
}

std::string Projects::getIdentifier() {    
    return this->identifier;
}

void Projects::setIdentifier(std::string value) {
    this->identifier = value;
}

std::string Projects::getTitle() {    
    return this->title;
}

void Projects::setTitle(std::string value) {
    this->title = value;
}

std::string Projects::getDescription() {    
    return this->description;
}

void Projects::setDescription(std::string value) {
    this->description = value;
}

std::string Projects::getWorkflowId() {    
    return this->workflowId;
}

void Projects::setWorkflowId(std::string value) {
    this->workflowId = value;
}

int Projects::getCreated() {    
    return this->created;
}

void Projects::setCreated(int value) {
    this->created = value;
}

int Projects::getModified() {    
    return this->modified;
}

void Projects::setModified(int value) {
    this->modified = value;
}

bool Projects::isDeleted() {    
    return this->deleted;
}

void Projects::setDeleted(bool value) {
    this->deleted = value;
}

