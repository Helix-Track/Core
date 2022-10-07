/*
    PermissionContexts.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "PermissionContexts.h"

std::string PermissionContexts::getId() {    
    return this->id;
}

void PermissionContexts::setId(std::string value) {
    this->id = value;
}

std::string PermissionContexts::getTitle() {    
    return this->title;
}

void PermissionContexts::setTitle(std::string value) {
    this->title = value;
}

std::string PermissionContexts::getDescription() {    
    return this->description;
}

void PermissionContexts::setDescription(std::string value) {
    this->description = value;
}

int PermissionContexts::getCreated() {    
    return this->created;
}

void PermissionContexts::setCreated(int value) {
    this->created = value;
}

int PermissionContexts::getModified() {    
    return this->modified;
}

void PermissionContexts::setModified(int value) {
    this->modified = value;
}

bool PermissionContexts::isDeleted() {    
    return this->deleted;
}

void PermissionContexts::setDeleted(bool value) {
    this->deleted = value;
}

