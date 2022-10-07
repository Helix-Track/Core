/*
    Tickets.cpp
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "Tickets.h"

std::string Tickets::getId() {    
    return this->id;
}

void Tickets::setId(std::string value) {
    this->id = value;
}

int Tickets::getTicketNumber() {    
    return this->ticketNumber;
}

void Tickets::setTicketNumber(int value) {
    this->ticketNumber = value;
}

int Tickets::getPosition() {    
    return this->position;
}

void Tickets::setPosition(int value) {
    this->position = value;
}

std::string Tickets::getTitle() {    
    return this->title;
}

void Tickets::setTitle(std::string value) {
    this->title = value;
}

std::string Tickets::getDescription() {    
    return this->description;
}

void Tickets::setDescription(std::string value) {
    this->description = value;
}

int Tickets::getCreated() {    
    return this->created;
}

void Tickets::setCreated(int value) {
    this->created = value;
}

int Tickets::getModified() {    
    return this->modified;
}

void Tickets::setModified(int value) {
    this->modified = value;
}

std::string Tickets::getTicketTypeId() {    
    return this->ticketTypeId;
}

void Tickets::setTicketTypeId(std::string value) {
    this->ticketTypeId = value;
}

std::string Tickets::getTicketStatusId() {    
    return this->ticketStatusId;
}

void Tickets::setTicketStatusId(std::string value) {
    this->ticketStatusId = value;
}

std::string Tickets::getProjectId() {    
    return this->projectId;
}

void Tickets::setProjectId(std::string value) {
    this->projectId = value;
}

std::string Tickets::getUserId() {    
    return this->userId;
}

void Tickets::setUserId(std::string value) {
    this->userId = value;
}

double Tickets::getEstimation() {    
    return this->estimation;
}

void Tickets::setEstimation(double value) {
    this->estimation = value;
}

int Tickets::getStoryPoints() {    
    return this->storyPoints;
}

void Tickets::setStoryPoints(int value) {
    this->storyPoints = value;
}

std::string Tickets::getCreator() {    
    return this->creator;
}

void Tickets::setCreator(std::string value) {
    this->creator = value;
}

bool Tickets::isDeleted() {    
    return this->deleted;
}

void Tickets::setDeleted(bool value) {
    this->deleted = value;
}

