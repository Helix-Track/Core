/*
    LabelLabelCategoryMappings.h
    Generated with 'sql2code' 0.0.1
    https://github.com/red-elf/SQL-to-Code
*/

#include "string"

class LabelLabelCategoryMappings {

private:
    std::string id;
    std::string labelId;
    std::string labelCategoryId;
    int created;
    int modified;
    bool deleted;

public:
    std::string getId();
    void setId(std::string value);
    std::string getLabelId();
    void setLabelId(std::string value);
    std::string getLabelCategoryId();
    void setLabelCategoryId(std::string value);
    int getCreated();
    void setCreated(int value);
    int getModified();
    void setModified(int value);
    bool isDeleted();
    void setDeleted(bool value);
};