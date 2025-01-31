# Medical Care App Dashboard Query

## Application Overview
Healthcare app for users with chronic conditions (e.g., rare allergies):
- Manages treatment plans with reminders for checkups/appointments
- Delivers relevant medical research updates and healthcare events

## Dashboard Features
1. Treatment Timeline
    - Recent completed activities with outcomes
    - Upcoming scheduled tasks
2. Personalized Article Feed
    - Ranked by user health segment relevance
    - Article engagement tracking (unread/viewed/read/saved)

## Technical Challenge
Complexity: Medium

Performance bottlenecks:
- Dashboard endpoint experiencing RPS scaling issues
- Complex relational data model with multiple joins
- Real-time personalization requirements

## Database Structure

```mermaid
erDiagram
    User ||--o{ UserSegment : has
    User ||--o{ CarePlan : receives
    User ||--o{ ReadArticle : reads
    
    UserSegment }|--|| SegmentType : "belongs to"
    
    Article ||--o{ ArticleSegment : tagged_with
    ArticleSegment }|--|| SegmentType : references
    
    CarePlan ||--o{ CarePlanSegment : targets
    CarePlanSegment }|--|| SegmentType : references
    
    Article ||--|{ ReadArticle : read_as

    CarePlan ||--|{ CarePlanStep : contains

    User {
        uuid id PK
        string name
        timestamp created_at
        timestamp updated_at
    }

    UserSegment {
        uuid id PK
        uuid user_id FK
        uuid segment_id FK
        float weight
        timestamp calculated_at
    }

    SegmentType {
        uuid id PK
        string name
        string description
        string medical_category
    }

    Article {
        uuid id PK
        string title
        text content
        string source
        string type "news or scientific"
        timestamp published_at
    }

    ArticleSegment {
        uuid id PK
        uuid article_id FK
        uuid segment_id FK
        float relevance_score
    }

    CarePlan {
        uuid id PK
        uuid user_id FK
        string title
        text description
        string status
        timestamp start_date
        timestamp end_date
    }

    CarePlanSegment {
        uuid id PK
        uuid care_plan_id FK
        uuid segment_id FK
        float importance_score
    }

    ReadArticle {
        uuid id PK
        uuid user_id FK
        uuid article_id FK
        timestamp read_at
        bool is_saved
    }

    CarePlanStep {
        uuid id PK
        uuid care_plan_id FK
        string type "appointment|test|upload|checkup"
        string title
        text description
        string status "pending|completed|overdue"
        timestamp due_date
        timestamp completed_at
        json metadata "doctor_info|test_type|required_docs"
        int order_number
    }
```