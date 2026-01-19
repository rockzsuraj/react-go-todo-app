import React from 'react';

export default function LoadingSkeleton() {
    return (
        <div className="container mt-4 animate-pulse">
            <div className="d-flex justify-content-between mb-4">
                <div className="bg-light rounded" style={{ width: '200px', height: '40px' }}></div>
                <div className="bg-light rounded" style={{ width: '100px', height: '40px' }}></div>
            </div>
            <div className="card mb-3">
                <div className="card-body">
                    <div className="bg-light rounded mb-2" style={{ width: '100%', height: '24px' }}></div>
                    <div className="bg-light rounded" style={{ width: '60%', height: '16px' }}></div>
                </div>
            </div>
            <div className="card mb-3">
                <div className="card-body">
                    <div className="bg-light rounded mb-2" style={{ width: '100%', height: '24px' }}></div>
                    <div className="bg-light rounded" style={{ width: '60%', height: '16px' }}></div>
                </div>
            </div>
            <div className="card mb-3">
                <div className="card-body">
                    <div className="bg-light rounded mb-2" style={{ width: '100%', height: '24px' }}></div>
                    <div className="bg-light rounded" style={{ width: '60%', height: '16px' }}></div>
                </div>
            </div>
        </div>
    );
}
