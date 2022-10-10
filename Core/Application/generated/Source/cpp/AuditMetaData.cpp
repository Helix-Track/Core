/*
    AuditMetaData.cpp
    Generated with 'sql2code' 0.0.2-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "AuditMetaData.h"

std::string AuditMetaData::getId() {    
    return this->id;
}

void AuditMetaData::setId(std::string &value) {
    this->id = value;
}

std::string AuditMetaData::getAuditId() {    
    return this->auditId;
}

void AuditMetaData::setAuditId(std::string &value) {
    this->auditId = value;
}

std::string AuditMetaData::getProperty() {    
    return this->property;
}

void AuditMetaData::setProperty(std::string &value) {
    this->property = value;
}

std::string AuditMetaData::getValue() {    
    return this->value;
}

void AuditMetaData::setValue(std::string &value) {
    this->value = value;
}

int AuditMetaData::getCreated() {    
    return this->created;
}

void AuditMetaData::setCreated(int &value) {
    this->created = value;
}

int AuditMetaData::getModified() {    
    return this->modified;
}

void AuditMetaData::setModified(int &value) {
    this->modified = value;
}

