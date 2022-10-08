/*
    PermissionContext.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "PermissionContext.h"

std::string PermissionContext::getId() {    
    return this->id;
}

void PermissionContext::setId(std::string value) {
    this->id = value;
}

std::string PermissionContext::getTitle() {    
    return this->title;
}

void PermissionContext::setTitle(std::string value) {
    this->title = value;
}

std::string PermissionContext::getDescription() {    
    return this->description;
}

void PermissionContext::setDescription(std::string value) {
    this->description = value;
}

int PermissionContext::getCreated() {    
    return this->created;
}

void PermissionContext::setCreated(int value) {
    this->created = value;
}

int PermissionContext::getModified() {    
    return this->modified;
}

void PermissionContext::setModified(int value) {
    this->modified = value;
}

bool PermissionContext::isDeleted() {    
    return this->deleted;
}

void PermissionContext::setDeleted(bool value) {
    this->deleted = value;
}

