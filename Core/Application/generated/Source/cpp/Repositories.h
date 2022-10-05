/*
    Repositories.h
    Generated with 'sql2code' 1.0.0-SNAPSHOT
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class Repositories {

private:
    std::string id;
    std::string repository;
    std::string description;
    std::string repositoryTypeId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getRepository();
    void setRepository(std::string value);
    std::string getDescription();
    void setDescription(std::string value);
    std::string getRepositoryTypeId();
    void setRepositoryTypeId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};