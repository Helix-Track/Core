/*
    Project.cpp
    Generated with 'sql2code' 0.0.2
    https://github.com/red-elf/SQL-to-Code
*/

#include "Project.h"

std::string Project::getId() {    
    return this->id;
}

void Project::setId(std::string &value) {
    this->id = value;
}

std::string Project::getIdentifier() {    
    return this->identifier;
}

void Project::setIdentifier(std::string &value) {
    this->identifier = value;
}

std::string Project::getTitle() {    
    return this->title;
}

void Project::setTitle(std::string &value) {
    this->title = value;
}

std::string Project::getDescription() {    
    return this->description;
}

void Project::setDescription(std::string &value) {
    this->description = value;
}

std::string Project::getWorkflowId() {    
    return this->workflowId;
}

void Project::setWorkflowId(std::string &value) {
    this->workflowId = value;
}

int Project::getCreated() {    
    return this->created;
}

void Project::setCreated(int &value) {
    this->created = value;
}

int Project::getModified() {    
    return this->modified;
}

void Project::setModified(int &value) {
    this->modified = value;
}

bool Project::isDeleted() {    
    return this->deleted;
}

void Project::setDeleted(bool &value) {
    this->deleted = value;
}

