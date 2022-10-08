/*
    CycleProjectMapping.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "CycleProjectMapping.h"

std::string CycleProjectMapping::getId() {    
    return this->id;
}

void CycleProjectMapping::setId(std::string value) {
    this->id = value;
}

std::string CycleProjectMapping::getCycleId() {    
    return this->cycleId;
}

void CycleProjectMapping::setCycleId(std::string value) {
    this->cycleId = value;
}

std::string CycleProjectMapping::getProjectId() {    
    return this->projectId;
}

void CycleProjectMapping::setProjectId(std::string value) {
    this->projectId = value;
}

int CycleProjectMapping::getCreated() {    
    return this->created;
}

void CycleProjectMapping::setCreated(int value) {
    this->created = value;
}

int CycleProjectMapping::getModified() {    
    return this->modified;
}

void CycleProjectMapping::setModified(int value) {
    this->modified = value;
}

bool CycleProjectMapping::isDeleted() {    
    return this->deleted;
}

void CycleProjectMapping::setDeleted(bool value) {
    this->deleted = value;
}

